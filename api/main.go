package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type BFClient struct {
	ctx         context.Context
	session     string
	redisClient *redis.Client
	buf         strings.Builder
}

func NewBF() *BFClient {
	return &BFClient{
		ctx:         context.Background(),
		session:     time.Now().String(),
		redisClient: redis.NewClient(&redis.Options{Addr: "redis:6379"}),
	}
}

func (bf *BFClient) Run(code string) (err error) {
	for i := 0; i < len(code); i++ {
		switch c := code[i]; c {
		case '>':
			err = bf.movePointer(1)
		case '<':
			err = bf.movePointer(-1)
		case '+':
			err = bf.add(1)
		case '-':
			err = bf.add(-1)
		case '[':
			if v, err := bf.value(); err != nil {
				return nil
			} else if v == 0 {
				// Find a matching bracket
				for loops := 1; loops > 0 && i < len(code); i++ {
					if code[i] == '[' {
						loops++
					} else if code[i] == ']' {
						loops--
					}
				}
			} else {
				err = bf.push(i - 1)
			}
		case ']':
			if v, err := bf.value(); err != nil {
				return err
			} else if v != 0 {
				i, err = bf.pop()
			}
		case '.':
			v, err := bf.value()
			if err != nil {
				return nil
			}
			bf.buf.WriteByte(byte(v))
		}
		if err != nil {
			log.Println("ERROR:", err)
			return err
		}
	}
	return nil
}

func (bf *BFClient) Output() string {
	return bf.buf.String()
}

func (bf *BFClient) movePointer(delta int) error {
	uri := "right"
	if delta < 0 {
		uri = "left"
	}
	req, err := http.NewRequest("GET", "http://ptr:8080/"+uri, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("X-Session", bf.session)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	} else if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	return nil
}

func (bf *BFClient) pointer() (int, error) {
	req, err := http.NewRequest("GET", "http://ptr:8080/", nil)
	if err != nil {
		return 0, nil
	}
	req.Header.Set("X-Session", bf.session)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	ptr, err := strconv.Atoi(string(b))
	if err != nil {
		return 0, err
	}
	return ptr, nil
}

func (bf *BFClient) value() (int, error) {
	ptr, err := bf.pointer()
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://mem:8080/%d", ptr), nil)
	if err != nil {
		return 0, nil
	}
	req.Header.Set("X-Session", bf.session)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	val, err := strconv.Atoi(string(b))
	return val, err
}

func (bf *BFClient) add(delta int) error {
	ptr, err := bf.pointer()
	if err != nil {
		return err
	}
	op := "inc"
	if delta < 0 {
		op = "dec"
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://mem:8080/%s/%d", op, ptr), nil)
	if err != nil {
		return nil
	}
	req.Header.Set("X-Session", bf.session)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	} else if res.StatusCode != http.StatusOK {
		return errors.New(res.Status)
	}
	return nil
}

func (bf *BFClient) push(pc int) error {
	return bf.redisClient.LPush(bf.ctx, bf.session+":stack", pc).Err()
}

func (bf *BFClient) pop() (int, error) {
	cmd := bf.redisClient.LPop(bf.ctx, bf.session+":stack")
	return cmd.Int()
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bf := NewBF()
		log.Println(r.FormValue("q"))
		err := bf.Run(r.FormValue("q"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprintf(w, "%s", bf.Output())
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
