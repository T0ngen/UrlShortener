package deletea

import (
	"errors"
	
	"io"
	"log/slog"
	"net/http"
	hashedapi "url-shortener/internal/hashedApi"

	"github.com/go-chi/render"
)


type Request struct{
	Alias string `json:"alias"`
}


type Delete interface{
	DeleteUrl(hashedapi string, alias string) (bool, error)
}


func New(log *slog.Logger, urlDeleter Delete)http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		hashToken := hashedapi.HashApi(token)


		var req Request
		
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// Такую ошибку встретим, если получили запрос с пустым телом.
			// Обработаем её отдельно
			log.Error("request body is empty")

			http.Error(w, "empty request", http.StatusInternalServerError)
			return
		}
		alias := req.Alias
		tr, err :=urlDeleter.DeleteUrl(hashToken, alias)
		


		
		if err != nil{
			http.Error(w, "Cant delete it", http.StatusInternalServerError)
			return
		}
		if tr{
			w.WriteHeader(http.StatusOK)
			return
		}else{
			http.Error(w, "Cant delete it", http.StatusInternalServerError)
			return
		}
	}
}