package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sarcb/catalyst-sp24/database/migrations"
	"github.com/sarcb/catalyst-sp24/generated/model"
	"github.com/sarcb/catalyst-sp24/test"
)

var (
	template1 = &model.TicketTemplateForm{
		Schema: migrations.DefaultTemplateSchema,
		Name:   "Template 1",
	}
	default1 = &model.TicketTemplateForm{
		Schema: migrations.DefaultTemplateSchema,
		Name:   "Default",
	}
)

func TestDatabase_TemplateCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		template *model.TicketTemplateForm
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Normal", args: args{template: template1}},
		{name: "Duplicate", args: args{template: default1}, wantErr: true},
		{name: "Nil template", args: args{}, wantErr: true},
		{name: "Template without fields", args: args{template: &model.TicketTemplateForm{}}, wantErr: true},
		{name: "Only name", args: args{template: &model.TicketTemplateForm{Name: "name"}}, wantErr: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, _, _, _, db, cleanup, err := test.DB(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()

			if _, err := db.TemplateCreate(test.Context(), tt.args.template); (err != nil) != tt.wantErr {
				t.Errorf("TemplateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDatabase_TemplateDelete(t *testing.T) {
	t.Parallel()

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Normal", args: args{"default"}},
		{name: "Not existing", args: args{"foobar"}, wantErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, _, _, _, db, cleanup, err := test.DB(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()

			if _, err := db.TemplateCreate(test.Context(), template1); err != nil {
				t.Errorf("TemplateCreate() error = %v", err)
			}

			if err := db.TemplateDelete(test.Context(), tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("TemplateDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDatabase_TemplateGet(t *testing.T) {
	t.Parallel()

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    *model.TicketTemplateResponse
		wantErr bool
	}{
		{name: "Normal", args: args{id: "default"}, want: &model.TicketTemplateResponse{ID: "default", Name: "Default", Schema: migrations.DefaultTemplateSchema}},
		{name: "Not existing", args: args{id: "foobar"}, wantErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, _, _, _, db, cleanup, err := test.DB(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()

			if _, err := db.TemplateCreate(test.Context(), template1); err != nil {
				t.Errorf("TemplateCreate() error = %v", err)
			}

			got, err := db.TemplateGet(test.Context(), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("TemplateGet() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if err != nil {
				return
			}

			assert.Equal(t, got, tt.want)
		})
	}
}

func TestDatabase_TemplateList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    []*model.TicketTemplateResponse
		wantErr bool
	}{
		{name: "Normal", want: []*model.TicketTemplateResponse{{ID: "default", Name: "Default", Schema: migrations.DefaultTemplateSchema}, {ID: "template-1", Name: template1.Name, Schema: template1.Schema}}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, _, _, _, db, cleanup, err := test.DB(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()

			if _, err := db.TemplateCreate(test.Context(), template1); err != nil {
				t.Errorf("TemplateCreate() error = %v", err)
			}

			got, err := db.TemplateList(test.Context())
			if (err != nil) != tt.wantErr {
				t.Errorf("TemplateList() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestDatabase_TemplateUpdate(t *testing.T) {
	t.Parallel()

	type args struct {
		id       string
		template *model.TicketTemplateForm
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Normal", args: args{"default", template1}},
		{name: "Not existing", args: args{"foobar", template1}, wantErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, _, _, _, db, cleanup, err := test.DB(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()

			if _, err := db.TemplateCreate(test.Context(), template1); err != nil {
				t.Errorf("TemplateCreate() error = %v", err)
			}

			if _, err := db.TemplateUpdate(test.Context(), tt.args.id, tt.args.template); (err != nil) != tt.wantErr {
				t.Errorf("TemplateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
