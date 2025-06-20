package zog

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Aviral-BFRS01281/zog/tutils"
	"github.com/Aviral-BFRS01281/zog/zconst"
	"github.com/stretchr/testify/assert"
)

// !STRUCTS

type User struct {
	Name string
}

type Team struct {
	Users []User
}

func TestSliceOfStructs(t *testing.T) {

	var userSchema = Struct(Shape{
		"name": String().Required(),
	})

	var teamSchema = Struct(Shape{
		"users": Slice(userSchema),
	})

	var data = map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"name": "Jane",
			},
			map[string]interface{}{
				"name": "John",
			},
		},
	}
	var team Team

	errsMap := teamSchema.Parse(data, &team)
	assert.Nil(t, errsMap)
	assert.Len(t, team.Users, 2)
	assert.Equal(t, team.Users[0].Name, "Jane")
	assert.Equal(t, team.Users[1].Name, "John")

	data = map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{},
			map[string]interface{}{},
		},
	}
	errsMap = teamSchema.Parse(data, &team)

	assert.Len(t, errsMap["users[0].name"], 1)
	assert.Len(t, errsMap["users[1].name"], 1)
	tutils.VerifyDefaultIssueMessagesMap(t, errsMap)
}

func TestSliceOptionalSlice(t *testing.T) {
	s := []string{}
	schema := Slice(String()) // should be optional by default

	errs := schema.Parse(nil, &s)
	assert.Nil(t, errs)
	assert.Len(t, s, 0)

	schema.Required().Optional() // should override required
	errs = schema.Parse(nil, &s)
	assert.Nil(t, errs)
	assert.Len(t, s, 0)
}

func TestSliceRequired(t *testing.T) {
	s := []string{}
	customMsg := "This slice is required and cannot be empty"
	schema := Slice(String()).Required(Message(customMsg))

	// Test with nil value
	errs := schema.Parse(nil, &s)
	assert.NotNil(t, errs)
	assert.Equal(t, customMsg, errs["$root"][0].Message)

	// Test with empty slice
	errs = schema.Parse([]string{}, &s)
	assert.Nil(t, errs)
	assert.Len(t, s, 0)

}
func TestSliceDefaultCoercing(t *testing.T) {
	s := []string{}
	schema := Slice(String())
	errs := schema.Parse("a", &s)
	assert.Nil(t, errs)
	assert.Len(t, s, 1)
	assert.Equal(t, s[0], "a")
}

func TestSliceDefault(t *testing.T) {
	schema := Slice(String()).Default([]string{"a", "b", "c"})
	s := []string{}
	err := schema.Parse(nil, &s)
	assert.Nil(t, err)
	assert.Len(t, s, 3)
	assert.Equal(t, s[0], "a")
	assert.Equal(t, s[1], "b")
	assert.Equal(t, s[2], "c")
}

func TestSlicePassSchema(t *testing.T) {

	s := []string{}
	schema := Slice(String().Required())

	errs := schema.Parse([]any{"a", "b", "c"}, &s)
	assert.Nil(t, errs)
	assert.Len(t, s, 3)
	assert.Equal(t, s[0], "a")
	assert.Equal(t, s[1], "b")
	assert.Equal(t, s[2], "c")
}

func TestSliceErrors(t *testing.T) {
	s := []string{}
	schema := Slice(String().Required().Min(2))

	errs := schema.Parse([]any{"a", "b"}, &s)
	assert.Len(t, errs, 3)
	assert.NotEmpty(t, errs["[0]"])
	assert.NotEmpty(t, errs["[1]"])
	assert.Empty(t, errs["[2]"])
	tutils.VerifyDefaultIssueMessagesMap(t, errs)
}

func TestSliceTransform(t *testing.T) {
	s := []string{}
	schema := Slice(String()).Transform(func(dataPtr any, ctx Ctx) error {
		s := dataPtr.(*[]string)
		for i := 0; i < len(*s); i++ {
			(*s)[i] = strings.ToUpper((*s)[i])
		}
		return nil
	})

	errs := schema.Parse([]string{"a", "b", "c"}, &s)

	assert.Nil(t, errs)
	assert.Len(t, s, 3)
	assert.Equal(t, []string{"A", "B", "C"}, s)
}

// VALIDATORS

func TestSliceLen(t *testing.T) {
	s := []string{}

	els := []string{"a", "b", "c", "d", "e"}
	schema := Slice(String().Required()).Len(2)
	errs := schema.Parse(els[:2], &s)
	assert.Len(t, s, 2)
	assert.Nil(t, errs)
	errs = schema.Parse(els[:1], &s)
	assert.NotEmpty(t, errs)
	tutils.VerifyDefaultIssueMessagesMap(t, errs)
	// min
	schema = Slice(String().Required()).Min(2)
	errs = schema.Parse(els[:4], &s)
	assert.Nil(t, errs)
	errs = schema.Parse(els[:1], &s)
	assert.NotEmpty(t, errs)
	tutils.VerifyDefaultIssueMessagesMap(t, errs)
	// max
	schema = Slice(String().Required()).Max(3)
	errs = schema.Parse(els[:1], &s)
	assert.Nil(t, errs)
	errs = schema.Parse(els[:4], &s)
	assert.NotNil(t, errs)
	tutils.VerifyDefaultIssueMessagesMap(t, errs)
}

func TestSliceContains(t *testing.T) {

	s := []string{}
	items := []string{"a", "b", "c"}

	schema := Slice(String()).Contains("a")
	errs := schema.Parse(items, &s)
	assert.Nil(t, errs)
	assert.Len(t, s, 3)

	schema = Slice(String()).Contains("d")
	errs = schema.Parse(items, &s)
	assert.NotEmpty(t, errs)
	tutils.VerifyDefaultIssueMessagesMap(t, errs)
}

func TestSliceCustomTest(t *testing.T) {
	input := []string{"abc", "defg", "hijkl"}
	s := []string{}
	schema := Slice(String()).TestFunc(func(val any, ctx Ctx) bool {
		// Custom test logic here
		x := val.(*[]string)
		return reflect.DeepEqual(input, *x)
	}, Message("custom"))
	errs := schema.Parse(input, &s)
	assert.Empty(t, errs)
	assert.Equal(t, input, s)
	errs = schema.Parse(input[1:], &s)
	assert.NotEmpty(t, errs)
	assert.Equal(t, "custom", errs["$root"][0].Message)
	// assert.Equal(t, "custom_test", errs["$root"][0].Code())
}

func TestSliceSchemaOption(t *testing.T) {
	s := Slice(String(), WithCoercer(func(original any) (value any, err error) {
		return []string{"coerced"}, nil
	}))

	var result []string
	err := s.Parse(123, &result)
	assert.Nil(t, err)
	assert.Equal(t, []string{"coerced"}, result)
}

func TestSliceGetType(t *testing.T) {
	s := Slice(String())
	assert.Equal(t, zconst.TypeSlice, s.getType())
}
