package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/kientink26/golang-started/pkg/forms"
	"github.com/kientink26/golang-started/pkg/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	threads, err := app.threads.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}
	for i := range threads {
		threads[i].User, err = app.users.Get(threads[i].User.ID)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	app.render(w, r, "home.page.tmpl", &templateData{
		Threads: threads,
	})
}

func (app *application) showThread(w http.ResponseWriter, r *http.Request) {
	// Pat doesn't strip the colon from the named capture key, so we need to
	// get the value of ":id" from the query string instead of "id".
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	s, err := app.threads.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	s.User, err = app.users.Get(s.User.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	s.Posts, err = app.posts.Latest(s.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	for i := range s.Posts {
		s.Posts[i].User, err = app.users.Get(s.Posts[i].User.ID)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	app.render(w, r, "showThread.page.tmpl", &templateData{
		Thread: s,
	})

}

func (app *application) createThreadForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "createThread.page.tmpl", &templateData{
		// Pass a new empty forms.Form object to the template.
		Form: forms.New(nil),
	})
}
func (app *application) createThread(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allowed", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	// First we call r.ParseForm() which adds any data in POST request bodies
	// to the r.PostForm map. If there are any errors, we use our app.ClientError helper to send
	// a 400 Bad Request response to the user.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := forms.New(r.PostForm)
	form.Required("topic")
	form.MaxLength("topic", 100)

	if !form.Valid() {
		app.render(w, r, "createThread.page.tmpl", &templateData{Form: form})
		return
	}

	userId := app.session.GetInt(r, "authenticatedUserID")
	id, err := app.threads.Insert(form.Get("topic"), userId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "Thread successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/thread/%d", id), http.StatusSeeOther)
}

func (app *application) createPostForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "createPost.page.tmpl", &templateData{
		// Pass a new empty forms.Form object to the template.
		Form: forms.New(nil),
	})
}

func (app *application) createPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allowed", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	threadId, err := strconv.Atoi(r.URL.Query().Get(":threadId"))
	if err != nil || threadId < 1 {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := forms.New(r.PostForm)
	form.Required("body")

	if !form.Valid() {
		app.render(w, r, "createPost.page.tmpl", &templateData{Form: form})
		return
	}

	userId := app.session.GetInt(r, "authenticatedUserID")
	_, err = app.posts.Insert(form.Get("body"), userId, threadId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "Post successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/thread/%d", threadId), http.StatusSeeOther)
}

func (app *application) deletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allowed", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	threadId, err := strconv.Atoi(r.URL.Query().Get(":threadId"))
	if err != nil || threadId < 1 {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	postId, err := strconv.Atoi(r.URL.Query().Get(":postId"))
	if err != nil || postId < 1 {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	userId := app.session.GetInt(r, "authenticatedUserID")
	post, err := app.posts.Get(postId)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}
	if post.User.ID != userId {
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	err = app.posts.Delete(postId, userId)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.session.Put(r, "flash", "Post successfully deleted!")

	http.Redirect(w, r, fmt.Sprintf("/thread/%d", threadId), http.StatusSeeOther)
}

func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}
func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	// Parse the form data.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// Validate the form contents using the form helper we made earlier.
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MaxLength("name", 255)
	form.MaxLength("email", 255)
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 10)
	// If there are any errors, redisplay the signup form.
	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}

	// Try to create a new user record in the database. If the email already exists
	// add an error message to the form and re-display it.
	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.Errors.Add("email", "Address is already in use")
			app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Otherwise add a confirmation flash message to the session confirming that
	// their signup worked and asking them to log in.
	app.session.Put(r, "flash", "Your signup was successful. Please log in.")
	// And redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}
func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &templateData{
		Form: forms.New(nil),
	})

}
func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	// Check whether the credentials are valid. If they're not, add a generic error
	// message to the form failures map and re-display the login page.
	form := forms.New(r.PostForm)
	id, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.Errors.Add("generic", "Email or Password is incorrect")
			app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Add the ID of the current user to the session, so that they are now 'logged
	// in'.
	app.session.Put(r, "authenticatedUserID", id)
	// Redirect the user to the
	// page they were originally trying to visit after logging in
	path := app.session.PopString(r, "redirectPathAfterLogin")
	if path != "" {
		http.Redirect(w, r, path, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	// Remove the authenticatedUserID from the session data so that the user is
	// 'logged out'.
	app.session.Remove(r, "authenticatedUserID")
	// Add a flash message to the session to confirm to the user that they've been
	// logged out.
	app.session.Put(r, "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userProfile(w http.ResponseWriter, r *http.Request) {
	id := app.session.GetInt(r, "authenticatedUserID")
	user, err := app.users.Get(id)
	if err != nil {
		app.serverError(w, err)
	}
	app.render(w, r, "profile.page.tmpl", &templateData{
		User: user,
	})
}
