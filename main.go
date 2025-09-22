package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	_ "image/png"
	"io"
	"log"
	"net"
	"net/http"
	"nonogram/board"
	"nonogram/encoding"
	"nonogram/hint"
	"nonogram/image"
	"nonogram/screen"
	"nonogram/solver"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

//go:embed static/index.html
var indexHTML []byte

var (
	ErrSolverTimeLimit = errors.New("solver time limit exceeded")
	ErrNoSolution      = errors.New("no solution found")

	mutex sync.Mutex
)

func decodeFromScreenshort(ctx context.Context, r io.Reader) (b *board.Board, h *hint.Hints, err error) {
	return screen.Decode(r)
}

func decodeFromText(ctx context.Context, r io.Reader) (b *board.Board, h *hint.Hints, err error) {
	return encoding.Decode(r)
}

func solve(ctx context.Context, b *board.Board, h *hint.Hints) (start time.Time, count uint64, decodedPNG, solvedPNG *bytes.Buffer, err error) {
	ctx, cancel := context.WithTimeoutCause(ctx, time.Minute, ErrSolverTimeLimit)
	defer cancel()

	decodedPNG = new(bytes.Buffer)
	err = image.Render(decodedPNG, b)
	if err != nil {
		return time.Time{}, 0, nil, nil, fmt.Errorf("failed to render image: %w", err)
	}

	solved, stats, err := solver.Solve(ctx, b, h)
	if cause := context.Cause(ctx); cause != nil {
		return time.Time{}, 0, decodedPNG, nil, cause
	}
	if err != nil {
		return time.Time{}, 0, decodedPNG, nil, err
	}
	if solved == nil {
		return time.Time{}, 0, decodedPNG, nil, ErrNoSolution
	}

	solvedPNG = new(bytes.Buffer)
	err = image.Render(solvedPNG, solved)
	if err != nil {
		return time.Time{}, 0, decodedPNG, nil, fmt.Errorf("failed to render image: %w", err)
	}

	return stats.Start, stats.Count, decodedPNG, solvedPNG, nil
}

func handleScreenshot(ctx context.Context, b *bot.Bot, update *models.Update) (*board.Board, *hint.Hints) {
	if update.Message.Document == nil || update.Message.Document.MimeType != "image/png" {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Please send a PNG screenshot of the Nonogram game. You have to send it as a file/document so that Telegram doesn't convert it to a JPG.",
		})
		return nil, nil
	}

	file, err := b.GetFile(ctx, &bot.GetFileParams{
		FileID: update.Message.Document.FileID,
	})
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      fmt.Sprintf("Failed to get file info:\n```%v```", err),
			ParseMode: models.ParseModeMarkdown,
		})
		return nil, nil
	}

	url := b.FileDownloadLink(file)

	res, err := http.Get(url)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      fmt.Sprintf("Failed to download file:\n```%v```", err),
			ParseMode: models.ParseModeMarkdown,
		})
		return nil, nil
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      fmt.Sprintf("Failed to download file:\n```%v```", err),
			ParseMode: models.ParseModeMarkdown,
		})
		return nil, nil
	}

	bd, h, err := decodeFromScreenshort(ctx, res.Body)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      fmt.Sprintf("Failed to decode screenshot:\n```%v```", err),
			ParseMode: models.ParseModeMarkdown,
		})
		return nil, nil
	}

	return bd, h
}

var firstLineRegexp = regexp.MustCompile(`(?m)^\d+ \d+$`)
var lastBoardSpecText string

func handleText(ctx context.Context, b *bot.Bot, update *models.Update) (*board.Board, *hint.Hints) {
	var text string
	if firstLineRegexp.MatchString(update.Message.Text) {
		lastBoardSpecText = update.Message.Text
		text = update.Message.Text
	} else {
		text = lastBoardSpecText + "\n" + update.Message.Text
	}
	bd, h, err := decodeFromText(ctx, bytes.NewBufferString(text))
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      fmt.Sprintf("Failed to decode text:\n```%v```", err),
			ParseMode: models.ParseModeMarkdown,
		})
		return nil, nil
	}
	return bd, h
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	mutex.Lock()
	defer mutex.Unlock()

	defer func() {
		if err := recover(); err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    update.Message.Chat.ID,
				Text:      fmt.Sprintf("An error occurred:\n```%v```", err),
				ParseMode: models.ParseModeMarkdown,
			})
		}
	}()

	if update.Message == nil {
		return
	}
	if update.Message.Chat.ID != 70260207 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "You are not authorized to use this bot.",
		})
		return
	}

	var bd *board.Board
	var h *hint.Hints
	if update.Message.Document != nil {
		bd, h = handleScreenshot(ctx, b, update)
	} else if update.Message.Text != "" {
		bd, h = handleText(ctx, b, update)
	}

	actionCtx, actionCtxCancel := context.WithCancel(ctx)
	defer actionCtxCancel()
	go func() {
		b.SendChatAction(ctx, &bot.SendChatActionParams{
			ChatID: update.Message.Chat.ID,
			Action: models.ChatActionTyping,
		})
		t := time.NewTicker(4 * time.Second)
		for {
			select {
			case <-actionCtx.Done():
				return
			case <-t.C:
				b.SendChatAction(ctx, &bot.SendChatActionParams{
					ChatID: update.Message.Chat.ID,
					Action: models.ChatActionTyping,
				})
			}
		}
	}()

	runtime.Gosched()
	start, count, _, solved, err := solve(ctx, bd, h)
	actionCtxCancel()
	runtime.Gosched()

	if errors.Is(err, ErrSolverTimeLimit) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "It took too long to solve the Nonogram. Please play a little more, figure out some more of the puzzle, and try again.",
		})
		return
	}
	if errors.Is(err, ErrNoSolution) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "No solution found for the Nonogram. Please try again.",
		})
		return
	}
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      fmt.Sprintf("Failed to solve Nonogram:\n```%v```", err),
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	took := time.Since(start)
	b.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:  update.Message.Chat.ID,
		Caption: fmt.Sprintf("Solved in %s after checking %d boards.", took, count),
		Photo: &models.InputFileUpload{
			Filename: "solved.png",
			Data:     solved,
		},
	})
}

func runTGBot(ctx context.Context) error {
	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(os.Getenv("TELEGRAM_BOT_KEY"), opts...)
	if err != nil {
		return err
	}

	b.Start(ctx)

	return nil
}

func runWebServer(ctx context.Context) error {
	listener, err := net.Listen("tcp", ":9999")
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	s := fuego.NewServer(
		fuego.WithListener(listener),
	)

	fuego.Get(s, "/", func(c fuego.ContextNoBody) (fuego.HTML, error) {
		c.SetHeader("Content-Type", "text/html; charset=utf-8")
		return fuego.HTML(indexHTML), nil
	})

	return s.Run()
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var wg sync.WaitGroup
	var err error

	wg.Add(2)

	go func() {
		e := runTGBot(ctx)
		if e != nil {
			err = e
		}
		wg.Done()
	}()

	go func() {
		e := runWebServer(ctx)
		if e != nil {
			err = e
		}
		wg.Done()
	}()

	wg.Wait()

	return err
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
