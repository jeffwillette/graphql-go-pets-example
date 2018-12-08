package main

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// User is the base user model to be used throughout the app
type User struct {
	gorm.Model
	Name string
	Pets []Pet `gorm:"foreignkey:OwnerID"`
}

func (db *DB) getUserPetIDs(ctx context.Context, userID uint) ([]int, error) {
	var ids []int
	err := db.DB.Where("owner_id = ?", userID).Find(&[]Pet{}).Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (db *DB) getUser(ctx context.Context, id uint) (*User, error) {
	var user User
	err := db.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserPets gets pets associated with the user
func (db *DB) GetUserPets(ctx context.Context, id uint) ([]Pet, error) {
	var u User
	u.ID = id

	var p []Pet
	err := db.DB.Model(&u).Association("Pets").Find(&p).Error
	if err != nil {
		return nil, err
	}

	return p, nil
}

// UserResolver contains the database and the user model to resolve against
type UserResolver struct {
	db *DB
	m  User
}

// ID resolves the user ID
func (u *UserResolver) ID(ctx context.Context) *graphql.ID {
	return gqlIDP(u.m.ID)
}

// Name resolves the Name field for User, it is all caps to avoid name clashes
func (u *UserResolver) Name(ctx context.Context) *string {
	return &u.m.Name
}

// Pets resolves the Pets field for User
func (u *UserResolver) Pets(ctx context.Context) (*[]*PetResolver, error) {
	pets, err := u.db.GetUserPets(ctx, u.m.ID)
	if err != nil {
		return nil, errors.Wrap(err, "Pets")
	}

	r := make([]*PetResolver, len(pets))
	for i := range pets {
		r[i] = &PetResolver{
			db: u.db,
			m:  pets[i],
		}
	}

	return &r, nil
}
