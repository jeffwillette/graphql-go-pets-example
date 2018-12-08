package main

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// Tag is the base type for a pet tag to be used by the db and gql
type Tag struct {
	gorm.Model
	Title string
	Pets  []Pet `gorm:"many2many:pet_tags"`
}

func (db *DB) getTagPets(ctx context.Context, t *Tag) ([]Pet, error) {
	var p []Pet
	err := db.DB.Model(t).Related(&p, "Pets").Error
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (db *DB) getTagBytTitle(ctx context.Context, title string) (*Tag, error) {
	var t Tag
	err := db.DB.Where("title = ?", title).First(&t).Error
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// TagResolver contains the db and the Tag model for resolving
type TagResolver struct {
	db *DB
	m  Tag
}

// ID resolves the ID for Tag
func (t *TagResolver) ID(ctx context.Context) *graphql.ID {
	return gqlIDP(t.m.ID)
}

// Title resolves the title field
func (t *TagResolver) Title(ctx context.Context) *string {
	return &t.m.Title
}

// Pets resolves the pets field
func (t *TagResolver) Pets(ctx context.Context) (*[]*PetResolver, error) {
	pets, err := t.db.getTagPets(ctx, &t.m)
	if err != nil {
		return nil, errors.Wrap(err, "Pets")
	}

	r := make([]*PetResolver, len(pets))
	for i := range pets {
		r[i] = &PetResolver{
			db: t.db,
			m:  pets[i],
		}
	}

	return &r, nil
}
