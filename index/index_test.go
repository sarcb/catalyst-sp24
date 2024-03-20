package index_test

import (
	"reflect"
	"testing"

	"github.com/sarcb/catalyst-sp24/generated/model"
	"github.com/sarcb/catalyst-sp24/test"
)

func TestIndex(t *testing.T) {
	t.Parallel()

	type args struct {
		term string
	}
	tests := []struct {
		name    string
		args    args
		wantIds []string
		wantErr bool
	}{
		{name: "Exists", args: args{"foo"}, wantIds: []string{"1"}},
		{name: "Not exists", args: args{"bar"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i, cleanup, err := test.Index(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()

			i.Index([]*model.TicketSimpleResponse{
				{ID: 0, Name: "bar"},
				{ID: 1, Name: "foo"},
			})

			gotIds, err := i.Search(tt.args.term)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(gotIds, tt.wantIds) {
				t.Errorf("Search() gotIds = %v, want %v", gotIds, tt.wantIds)
			}
		})
	}
}

func TestIndex_Truncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "Truncate"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			i, cleanup, err := test.Index(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()

			i.Index([]*model.TicketSimpleResponse{
				{ID: 0, Name: "bar"},
				{ID: 1, Name: "foo"},
			})

			if err := i.Truncate(); (err != nil) != tt.wantErr {
				t.Errorf("Truncate() error = %v, wantErr %v", err, tt.wantErr)
			}

			ids, err := i.Search("foo")
			if err != nil {
				t.Fatal(err)
			}

			if ids != nil {
				t.Fatal("should return no results")
			}
		})
	}
}
