package redisx

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	addr        string
	password    string
	db          int
	dialTimeout time.Duration
}

func NewClient(addr, password string, db int) *Client {
	return &Client{
		addr:        addr,
		password:    password,
		db:          db,
		dialTimeout: 5 * time.Second,
	}
}

func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.do(ctx, []string{"PING"})
	if err != nil {
		return err
	}
	pong, ok := resp.(string)
	if !ok || pong != "PONG" {
		return fmt.Errorf("unexpected ping response: %v", resp)
	}
	return nil
}

func (c *Client) EvalInt(ctx context.Context, script string, keys []string, args []string) (int64, error) {
	command := make([]string, 0, 3+len(keys)+len(args))
	command = append(command, "EVAL", script, strconv.Itoa(len(keys)))
	command = append(command, keys...)
	command = append(command, args...)

	resp, err := c.do(ctx, command)
	if err != nil {
		return 0, err
	}
	value, ok := resp.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected eval response type: %T", resp)
	}
	return value, nil
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) do(ctx context.Context, args []string) (any, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := c.applyDeadline(ctx, conn); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	if err := c.authenticate(conn, reader); err != nil {
		return nil, err
	}
	if err := c.selectDB(conn, reader); err != nil {
		return nil, err
	}
	if err := writeArray(conn, args); err != nil {
		return nil, err
	}

	return readReply(reader)
}

func (c *Client) dial(ctx context.Context) (net.Conn, error) {
	dialer := net.Dialer{Timeout: c.dialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return nil, fmt.Errorf("dial redis %s: %w", c.addr, err)
	}
	return conn, nil
}

func (c *Client) applyDeadline(ctx context.Context, conn net.Conn) error {
	if deadline, ok := ctx.Deadline(); ok {
		return conn.SetDeadline(deadline)
	}
	return conn.SetDeadline(time.Now().Add(c.dialTimeout))
}

func (c *Client) authenticate(conn net.Conn, reader *bufio.Reader) error {
	if c.password == "" {
		return nil
	}
	if err := writeArray(conn, []string{"AUTH", c.password}); err != nil {
		return err
	}
	resp, err := readReply(reader)
	if err != nil {
		return err
	}
	if ok, okType := resp.(string); !okType || ok != "OK" {
		return fmt.Errorf("unexpected auth response: %v", resp)
	}
	return nil
}

func (c *Client) selectDB(conn net.Conn, reader *bufio.Reader) error {
	if c.db == 0 {
		return nil
	}
	if err := writeArray(conn, []string{"SELECT", strconv.Itoa(c.db)}); err != nil {
		return err
	}
	resp, err := readReply(reader)
	if err != nil {
		return err
	}
	if ok, okType := resp.(string); !okType || ok != "OK" {
		return fmt.Errorf("unexpected select response: %v", resp)
	}
	return nil
}

func writeArray(conn net.Conn, args []string) error {
	if _, err := fmt.Fprintf(conn, "*%d\r\n", len(args)); err != nil {
		return err
	}
	for _, arg := range args {
		if _, err := fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(arg), arg); err != nil {
			return err
		}
	}
	return nil
}

func readReply(reader *bufio.Reader) (any, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	switch prefix {
	case '+':
		return line, nil
	case '-':
		return nil, errors.New(line)
	case ':':
		value, convErr := strconv.ParseInt(line, 10, 64)
		if convErr != nil {
			return nil, fmt.Errorf("parse redis integer: %w", convErr)
		}
		return value, nil
	case '$':
		size, convErr := strconv.Atoi(line)
		if convErr != nil {
			return nil, fmt.Errorf("parse redis bulk size: %w", convErr)
		}
		if size == -1 {
			return "", nil
		}
		buf := make([]byte, size+2)
		if _, err := io.ReadFull(reader, buf); err != nil {
			return nil, err
		}
		return string(buf[:size]), nil
	default:
		return nil, fmt.Errorf("unsupported redis reply prefix: %q", prefix)
	}
}
