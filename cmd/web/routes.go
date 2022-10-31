package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	mux := pat.New()
	dynamicMiddleware := alice.New(app.session.Enable, app.authenticate)
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Get("/thread/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createThreadForm))
	mux.Post("/thread/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createThread))
	mux.Get("/thread/:id", dynamicMiddleware.ThenFunc(app.showThread))
	mux.Get("/thread/:threadId/post/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createPostForm))
	mux.Post("/thread/:threadId/post/create", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.createPost))
	mux.Post("/thread/:threadId/post/:postId/delete", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.deletePost))

	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	mux.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.logoutUser))
	mux.Get("/user/profile", dynamicMiddleware.Append(app.requireAuthentication).ThenFunc(app.userProfile))

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(mux)
}
