package controllers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/hellerox/lenselocked/context"

	"github.com/hellerox/lenselocked/models"
	"github.com/hellerox/lenselocked/views"
)

const (
	IndexGalleries = "index_galleries"
	ShowGallery    = "show_gallery"
	EditGallery    = "edit_gallery"
)

func NewGalleries(gs models.GalleryService, r *mux.Router) *Galleries {
	return &Galleries{
		New:       views.NewView("bootstrap", "galleries/new"),
		ShowView:  views.NewView("bootstrap", "galleries/show"),
		EditView:  views.NewView("bootstrap", "galleries/edit"),
		IndexView: views.NewView("bootstrap", "galleries/index"),
		gs:        gs,
		r:         r,
	}
}

type Galleries struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	gs        models.GalleryService
	r         *mux.Router
}

type GalleryForm struct {
	Title string `schema:"title"`
}

func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}

	// This is our new code
	user := context.User(r.Context())
	// Then we update how we build the Gallery model
	gallery := models.Gallery{
		Title:  form.Title,
		UserID: user.ID,
	}

	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}

	// Then add the following code to create a URL using the
	// router and the named ShowGallery route.
	url, err := g.r.Get(EditGallery).URL("id",
		strconv.Itoa(int(gallery.ID)))
	// Check for errors creating the URL
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	// If no errors, use the URL we just created and redirect
	// to the path portion of that URL. We don't need the
	// entire URL because your application might be hosted at
	// localhost:3000, or it might be at lenslocked.com. By
	// only using the path our code is agnostic to that detail.
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// GET /galleries/:id
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		// The galleryByID method will already render the error
		// for us, so we just need to return here.
		return
	}
	var vd views.Data
	vd.Yield = gallery
	g.ShowView.Render(w, r, vd)
}

func (g *Galleries) galleryByID(w http.ResponseWriter,
	r *http.Request) (*models.Gallery, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}
	gallery, err := g.gs.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found", http.StatusNotFound)
		default:
			http.Error(w, "Whoops! Something went wrong.", http.StatusInternalServerError)
		}
		return nil, err
	}
	return gallery, nil
}

func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		// The galleryByID method will already render the error
		// for us, so we just need to return here.
		return
	}
	// A user needs logged in to access this page, so we can
	// assume that the RequireUser middleware has run and
	// set the user for us in the request context.
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You do not have permission to edit "+
			"this gallery", http.StatusForbidden)
		return
	}

	var vd views.Data
	vd.Yield = gallery
	g.EditView.Render(w, r, vd)
}

// POST /galleries/:id/update
func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	var vd views.Data
	vd.Yield = gallery
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		// If there is an error we are going to render the
		// EditView again with an alert message.
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	gallery.Title = form.Title
	gallery.Title = form.Title
	err = g.gs.Update(gallery)
	// If there is an error our alert will be an error. Otherwise
	// we will still render an alert, but instead it will be
	// a success message.
	if err != nil {
		vd.SetAlert(err)
	} else {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlSuccess,
			Message: "Gallery successfully updated!",
		}
	}
	// Error or not, we are going to render the EditView with
	// our updated information.
	g.EditView.Render(w, r, vd)
}

func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	// Lookup the gallery using the galleryByID we wrote earlier
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		// If there is an error the galleryByID will have rendered
		// it for us already.
		return
	}
	// We also need to retrieve the user and verify they have
	// permission to delete this gallery. This means we will
	// need to use the RequireUser middleware on any routes
	// mapped to this method.
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You do not have permission to edit "+
			"this gallery", http.StatusForbidden)
		return
	}

	var vd views.Data
	err = g.gs.Delete(gallery.ID)
	if err != nil {
		// If there is an error we want to set an alert and
		// render the edit page with the error. We also need
		// to set the Yield to gallery so that the EditView
		// is rendered correctly.
		vd.SetAlert(err)
		vd.Yield = gallery
		g.EditView.Render(w, r, vd)
		return
	}

	url, err := g.r.Get(IndexGalleries).URL()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.Redirect(w, r, url.Path, http.StatusFound)
}

// GET /galleries
func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	galleries, err := g.gs.ByUserID(user.ID)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	var vd views.Data
	vd.Yield = galleries
	g.IndexView.Render(w, r, vd)
}
