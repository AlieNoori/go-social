package main

import (
	"net/http"
	"strconv"

	"github.com/AlieNoori/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtxKey userKey = "user"

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil || userId < 1 {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()

	user, err := app.getUser(ctx, userId)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := app.writeResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User followed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromCtx(r)
	followedId, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	if err := app.store.Followers.Follow(r.Context(), followerUser.ID, followedId); err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.writeResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

// UnfollowUser gdoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User unfollowed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/unfollow [put]
func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromCtx(r)
	unfollowedId, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	if err := app.store.Followers.Unfollow(r.Context(), followerUser.ID, unfollowedId); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.writeResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

// ActivateUser godoc
//
//	@Summary		Activates/Register a user
//	@Description	Activates/Register a user by invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	if err := app.store.Users.Activate(r.Context(), token); err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return

	}

	if err := app.writeResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

// func (app *application) userContextMiaddleWare(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		idParam := chi.URLParam(r, "userID")
// 		postID, err := strconv.Atoi(idParam)
// 		if err != nil {
// 			app.badRequestResponse(w, r, err)
// 			return
// 		}
//
// 		ctx := r.Context()
//
// 		user, err := app.getUser(ctx, postID)
// 		if err != nil {
// 			switch err {
// 			case store.ErrNotFound:
// 				app.notFoundResponse(w, r, err)
// 			default:
// 				app.internalServerError(w, r, err)
// 			}
// 		}
//
// 		ctx = context.WithValue(r.Context(), userCtxKey, user)
//
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

func getUserFromCtx(r *http.Request) *store.User {
	post := r.Context().Value(userCtxKey).(*store.User)

	return post
}
