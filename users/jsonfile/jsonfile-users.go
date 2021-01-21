package jsonfile

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/go-msvc/japp/users"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
)

func Load(fn string) (users.IUsers, error) {
	f, err := os.Open(fn)
	if err == nil {
		defer f.Close()
		users := &jsonfileUsers{
			fn:    fn,
			users: map[string]*user{},
		}
		if err := json.NewDecoder(f).Decode(&users.users); err != nil {
			if err != io.EOF {
				return nil, errors.Wrapf(err, "failed to read JSON from file %s", fn)
			}
		}
		return users, nil
	}

	//cannot open file - create new file
	f, err = os.Create(fn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create users file %s", fn)
	}
	f.Close()
	users := &jsonfileUsers{
		fn:    fn,
		users: map[string]*user{},
	}
	jsonUsers, _ := json.Marshal(users.users)
	f.Write(jsonUsers)
	return users, nil
}

//implements users.IUsers
type jsonfileUsers struct {
	sync.Mutex
	fn    string
	users map[string]*user
}

func (uu *jsonfileUsers) New() (users.IUser, error) {
	uu.Lock()
	defer uu.Unlock()
	u := &user{
		id:   uuid.NewV1().String(),
		data: map[string]interface{}{},
	}
	uu.users[u.id] = u
	return u, nil
}

func (uu *jsonfileUsers) Get(id string) users.IUser {
	uu.Lock()
	defer uu.Unlock()
	if u, ok := uu.users[id]; ok {
		return u
	}
	return nil
}

func (uu *jsonfileUsers) update() error {
	f, err := os.Open(uu.fn)
	if err != nil {
		f, err = os.Create(uu.fn)
		if err != nil {
			return errors.Wrapf(err, "cannot create file %s", uu.fn)
		}
	}
	defer f.Close()
	jsonUsers, _ := json.Marshal(uu.users)
	if _, err = f.Write(jsonUsers); err != nil {
		return errors.Wrapf(err, "failed to write users to file %s", uu.fn)
	}
	return nil
}

//implements users.IUser
type user struct {
	id   string
	data map[string]interface{}
}

func (u user) ID() string {
	return u.id
}

func (u user) Data() map[string]interface{} {
	return u.data
}
