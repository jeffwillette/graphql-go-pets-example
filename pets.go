package main

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// Pet is the base type for pets to be used by the db and gql
type Pet struct {
	gorm.Model
	OwnerID uint
	Name    string
	Tags    []Tag `gorm:"many2many:pet_tags"`
}

// GetPet should authorize the user in ctx and return a pet or error
func (db *DB) getPet(ctx context.Context, id uint) (*Pet, error) {
	var p Pet
	err := db.DB.First(&p, id).Error
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (db *DB) getPetOwner(ctx context.Context, id int32) (*User, error) {
	var u User
	err := db.DB.First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *DB) getPetTags(ctx context.Context, p *Pet) ([]Tag, error) {
	var t []Tag
	err := db.DB.Model(p).Related(&t, "Tags").Error
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (db *DB) getPetsByID(ctx context.Context, ids []int, from, to int) ([]Pet, error) {
	var p []Pet
	err := db.DB.Where("id in (?)", ids[from:to]).Find(&p).Error
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (db *DB) updatePet(ctx context.Context, args *petInput) (*Pet, error) {
	// get the pet to be updated from the db
	var p Pet
	err := db.DB.First(&p, args.ID).Error
	if err != nil {
		return nil, err
	}

	// so the pointer dereference is safe
	if args.TagIDs == nil {
		return nil, errors.Wrap(err, "UpdatePet")
	}

	// if there are tags to be updated, go through that process
	var newTags []Tag
	if len(*args.TagIDs) > 0 {
		err = db.DB.Where("id in (?)", args.TagIDs).Find(&newTags).Error
		if err != nil {
			return nil, err
		}

		// replace the old tag set with the new one
		err = db.DB.Model(&p).Association("Tags").Replace(newTags).Error
		if err != nil {
			return nil, err
		}
	}

	updated := Pet{
		Name:    args.Name,
		OwnerID: uint(args.OwnerID),
	}

	err = db.DB.Model(&p).Updates(updated).Error
	if err != nil {
		return nil, err
	}

	err = db.DB.First(&p, args.ID).Error
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (db *DB) deletePet(ctx context.Context, userID, petID uint) (*bool, error) {
	// make sure the record exist
	var p Pet
	err := db.DB.First(&p, petID).Error
	if err != nil {
		return nil, err
	}

	// delete tags
	err = db.DB.Model(&p).Association("Tags").Clear().Error
	if err != nil {
		return nil, err
	}

	// delete record
	err = db.DB.Delete(&p).Error
	if err != nil {
		return nil, err
	}

	return boolP(true), err
}

func (db *DB) addPet(ctx context.Context, input petInput) (*Pet, error) {
	// get the M2M relation tags from the DB and put them in the pet to be saved
	var t []Tag
	err := db.DB.Where("id in (?)", input.TagIDs).Find(&t).Error
	if err != nil {
		return nil, err
	}

	pet := Pet{
		Name:    input.Name,
		OwnerID: uint(input.OwnerID),
		Tags:    t,
	}

	err = db.DB.Create(&pet).Error
	if err != nil {
		return nil, err
	}

	return &pet, nil
}

// PetResolver contains the DB and the model for resolving
type PetResolver struct {
	db *DB
	m  Pet
}

// ID resolves the ID field for Pet
func (p *PetResolver) ID(ctx context.Context) *graphql.ID {
	return gqlIDP(p.m.ID)
}

// Owner resolves the owner field for Pet
func (p *PetResolver) Owner(ctx context.Context) (*UserResolver, error) {
	user, err := p.db.getPetOwner(ctx, int32(p.m.OwnerID))
	if err != nil {
		return nil, errors.Wrap(err, "Owner")
	}

	r := UserResolver{
		db: p.db,
		m:  *user,
	}

	return &r, nil
}

// Name resolves the name field for Pet
func (p *PetResolver) Name(ctx context.Context) *string {
	return &p.m.Name
}

// Tags resolves the pet tags
func (p *PetResolver) Tags(ctx context.Context) (*[]*TagResolver, error) {
	tags, err := p.db.getPetTags(ctx, &p.m)
	if err != nil {
		return nil, errors.Wrap(err, "Tags")
	}

	r := make([]*TagResolver, len(tags))
	for i := range tags {
		r[i] = &TagResolver{
			db: p.db,
			m:  tags[i],
		}
	}

	return &r, nil
}
