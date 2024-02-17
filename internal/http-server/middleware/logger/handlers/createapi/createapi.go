package createapi

import (
	
	"encoding/json"
	"log/slog"
	"net/http"
	"url-shortener/internal/generateapi"
	hashedapi "url-shortener/internal/hashedApi"
)

type Response struct{
	Status string `json:"status"`
	ApiToken string `json:"api_token"`
	

}

type CreateApi interface {
	AddNewAPIToDb(api string) error
}

func New(log *slog.Logger, cr CreateApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newApi, err := generateapi.GenerateAPIKey()
		
		if err != nil{
			log.Info("mistake with API creation")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
		}
		hashedApi := hashedapi.HashApi(newApi)
		err = cr.AddNewAPIToDb(hashedApi)
		if err != nil{
			log.Info("mistake with API creation")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
            return

		}
		response := Response{Status: "SUCCESS", ApiToken: newApi}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			log.Info("error encoding JSON")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write(jsonResponse)
    }
}

		




