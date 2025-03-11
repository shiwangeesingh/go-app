package users

import 	"github.com/go-chi/chi/v5"

func Routes() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/register", CreateUser)
	r.Post("/login", AuthenticateUser)
	return r
}
