package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/ncostamagna/g_ms_enrollment_ex/internal/enrollment"
	"github.com/ncostamagna/g_ms_enrollment_ex/pkg/bootstrap"
	"github.com/ncostamagna/g_ms_enrollment_ex/pkg/handler"

	courseSdk "github.com/ncostamagna/g_sdk_ex/course"
	userSdk "github.com/ncostamagna/g_sdk_ex/user"
)

func main() {

	_ = godotenv.Load()
	l := bootstrap.InitLogger()
	db, err := bootstrap.DBConnection()
	if err != nil {
		l.Fatal(err)
	}

	ctx := context.Background()

	courseTrans := courseSdk.NewHttpClient(os.Getenv("API_COURSE_URL"), "")
	userTrans := userSdk.NewHttpClient(os.Getenv("API_USER_URL"), "")

	enrollRepo := enrollment.NewRepo(db, l)
	enrollSrv := enrollment.NewService(l, userTrans, courseTrans, enrollRepo)
	h := handler.NewEnrollmentHTTPServer(ctx, enrollment.MakeEndpoints(enrollSrv))
	port := os.Getenv("PORT")
	address := fmt.Sprintf("127.0.0.1:%s", port)
	srv := &http.Server{
		Handler:      accessControl(h),
		Addr:         address,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  4 * time.Second,
	}

	errCh := make(chan error)

	go func() {
		l.Println("listen in ", address)
		errCh <- srv.ListenAndServe()
	}()

	err = <-errCh
	if err != nil {
		log.Fatal(err)
	}

}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS, HEAD")
		w.Header().Set("Access-Control-Allow-Headers", "Accept,Authorization,Cache-Control,Content-Type,DNT,If-Modified-Since,Keep-Alive,Origin,User-Agent,X-Requested-With")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
