package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
)

// Resolver is the root resolver
type Resolver struct {
	db *DB
}

// GetUser resolves the getUser query
func (r *Resolver) GetUser(ctx context.Context, args struct{ ID graphql.ID }) (*UserResolver, error) {
	id, err := gqlIDToUint(args.ID)
	if err != nil {
		return nil, errors.Wrap(err, "GetPet")
	}

	user, err := r.db.getUser(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "GetUser")
	}

	s := UserResolver{
		db: r.db,
		m:  *user,
	}

	return &s, nil
}

// GetPet resolves the getPet query
func (r *Resolver) GetPet(ctx context.Context, args struct{ ID graphql.ID }) (*PetResolver, error) {
	id, err := gqlIDToUint(args.ID)
	if err != nil {
		return nil, errors.Wrap(err, "GetPet")
	}

	pet, err := r.db.getPet(ctx, id)
	if err != nil {
		return nil, err
	}

	s := PetResolver{
		db: r.db,
		m:  *pet,
	}

	return &s, nil
}

// GetTag resolves the getTag query
func (r *Resolver) GetTag(ctx context.Context, args struct{ Title string }) (*TagResolver, error) {
	tag, err := r.db.getTagBytTitle(ctx, args.Title)
	if err != nil {
		return nil, errors.Wrap(err, "GetTag")
	}

	s := TagResolver{
		db: r.db,
		m:  *tag,
	}

	return &s, nil
}

// petInput has everything needed to do adds and updates on a pet
type petInput struct {
	ID      *graphql.ID
	OwnerID int32
	Name    string
	TagIDs  *[]*int32
}

// AddPet Resolves the addPet mutation
func (r *Resolver) AddPet(ctx context.Context, args struct{ Pet petInput }) (*PetResolver, error) {
	pet, err := r.db.addPet(ctx, args.Pet)
	if err != nil {
		return nil, errors.Wrap(err, "AddPet")
	}

	s := PetResolver{
		db: r.db,
		m:  *pet,
	}

	return &s, nil
}

// UpdatePet takes care of updating any field on the pet
func (r *Resolver) UpdatePet(ctx context.Context, args struct{ Pet petInput }) (*PetResolver, error) {
	pet, err := r.db.updatePet(ctx, &args.Pet)
	if err != nil {
		return nil, errors.Wrap(err, "UpdatePet")
	}

	s := PetResolver{
		db: r.db,
		m:  *pet,
	}

	return &s, nil
}

// DeletePet takes care of deleting a pet record
func (r *Resolver) DeletePet(ctx context.Context, args struct{ UserID, PetID graphql.ID }) (*bool, error) {
	petID, err := gqlIDToUint(args.PetID)
	if err != nil {
		return nil, errors.Wrap(err, "DeletePet")
	}

	userID, err := gqlIDToUint(args.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "DeletePet")
	}

	return r.db.deletePet(ctx, userID, petID)
}

// encode cursor encodes the cursot position in base64
func encodeCursor(i int) graphql.ID {
	return graphql.ID(base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("cursor%d", i))))
}

// decode cursor decodes the base 64 encoded cursor and resturns the integer
func decodeCursor(s string) (int, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return 0, err
	}

	i, err := strconv.Atoi(strings.TrimPrefix(string(b), "cursor"))
	if err != nil {
		return 0, err
	}

	return i, nil
}
