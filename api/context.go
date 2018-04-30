package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dutchcoders/slackarchive/api/errors"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions" // /tree/ptrs
)

func init() {
}

type AfterFunc func()

type Context struct {
	db *database

	w           http.ResponseWriter
	r           *http.Request
	afterFn     AfterFunc
	Vars        map[string]string
	bodyWritten bool
	store       *sessions.CookieStore
}

type ContextFunc func(*Context) error

func (api *api) ContextHandlerFunc(h ContextFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := api.session.Copy()
		defer session.Close()

		ctx := Context{
			db: Database(session),

			r: r,
			w: w,
			afterFn: func() {
			},
			store:       api.store,
			bodyWritten: false,
			Vars:        mux.Vars(r),
		}

		var err error
		defer func() {
			if err == nil {
				if ctx.bodyWritten {
				} else {
					w.WriteHeader(http.StatusNoContent)
				}
				return
			}

			log.Error(err.Error())

			switch err.(type) {
			case *mysql.MySQLError:
				// or should we do this in the api call itself, instead of generic errors?
				driverErr := err.(*mysql.MySQLError)
				switch driverErr.Number {
				case 1062:
					w.WriteHeader(ErrDatabaseAlreadyExists.Code())
					json.NewEncoder(w).Encode(ErrDatabaseAlreadyExists)
				default:
					w.WriteHeader(ErrDatabaseOther.Code())
					json.NewEncoder(w).Encode(ErrDatabaseOther)
				}
			case errors.APIError:
				w.WriteHeader(err.(errors.APIError).Code())
				json.NewEncoder(w).Encode(err)
			default:
				http.Error(w, err.Error(), 500)
			}
		}()

		defer func() {
			ctx.afterFn()
			return
		}()

		err = h(&ctx)
		return
	}
}

func (ctx *Context) token() (token string, ok bool) {
	auth := ctx.r.Header.Get("Authorization")
	if auth == "" {
		return "", false
	}

	f := strings.Fields(auth)
	if len(f) != 2 || f[0] != "Token" {
		return "", false
	}

	return f[1], true
}

/*
func (ctx *Context) GetUser(user *model.User) error {
	token, ok := ctx.token()
	if !ok {
		return ErrNotAuthorized
	}

	if err := ctx.tx.Getx(user, model.QueryUserByTokenAndTokenType(token, model.TokenTypeLogin)); err == sql.ErrNoRows {
		return ErrNotFound
	} else {
		return err
	}
}*/

func (ctx *Context) Read(o interface{}) error {
	err := json.NewDecoder(ctx.r.Body).Decode(o)
	return err
}

func (ctx *Context) Write(o interface{}) error {
	ctx.w.WriteHeader(http.StatusOK)
	ctx.w.Header().Add("Content-Type", "application/json")
	ctx.bodyWritten = true
	err := json.NewEncoder(ctx.w).Encode(o)
	return err
}
