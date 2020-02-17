package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/hellerox/lenselocked/rand"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/hellerox/lenselocked/controllers"
	"github.com/hellerox/lenselocked/middleware"
	"github.com/hellerox/lenselocked/models"
)

func main() {
	boolPtr := flag.Bool("prod", false, "Provide this flag "+
		"in production. This ensures that a .config file is "+
		"provided before the application starts.")
	flag.Parse()
	// boolPtr is a pointer to a boolean, so we need to use
	// *boolPtr to get the boolean value and pass it into our
	// LoadConfig function
	cfg := LoadConfig(*boolPtr)
	dbCfg := cfg.Database
	services, err := models.NewServices(
		models.WithGorm(dbCfg.Dialect(), dbCfg.ConnectionInfo()),
		models.WithLogMode(!cfg.IsProd()),
		// We want each of these services, but if we didn't need
		// one of them we could possibly skip that config func
		models.WithUser(cfg.Pepper, cfg.HMACKey),
		models.WithGallery(),
		models.WithImage(),
	)

	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()

	r := mux.NewRouter()
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User)
	galleriesC := controllers.NewGalleries(services.Gallery, services.Image, r)

	userMw := middleware.User{
		UserService: services.User,
	}
	requireUserMw := middleware.RequireUser{}
	// The line above is the same as:
	//   var requireUserMw middleware.RequireUser

	b, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}
	// Use the config's IsProd method instead
	csrfMw := csrf.Protect(b, csrf.Secure(cfg.IsProd()))

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")
	// Gallery routes
	r.Handle("/galleries",
		requireUserMw.ApplyFn(galleriesC.Index)).
		Methods("GET").
		Name(controllers.IndexGalleries)
	r.Handle("/galleries/new",
		requireUserMw.Apply(galleriesC.New)).
		Methods("GET")
	r.Handle("/galleries",
		requireUserMw.ApplyFn(galleriesC.Create)).
		Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}",
		galleriesC.Show).
		Methods("GET").
		Name(controllers.ShowGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit",
		requireUserMw.ApplyFn(galleriesC.Edit)).
		Methods("GET").
		Name(controllers.EditGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/update",
		requireUserMw.ApplyFn(galleriesC.Update)).
		Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete",
		requireUserMw.ApplyFn(galleriesC.Delete)).
		Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images",
		requireUserMw.ApplyFn(galleriesC.ImageUpload)).
		Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete",
		requireUserMw.ApplyFn(galleriesC.ImageDelete)).
		Methods("POST")
	// Assets
	assetHandler := http.FileServer(http.Dir("./assets/"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)

	// Image routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	// Our port is not provided via config, so we need to
	// update the last bit of our main function.
	fmt.Printf("Starting the server on :%d...\n", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port),
		csrfMw(userMw.Apply(r)))
}
