package etcd

import (
	"errors"
	"fmt"

	"code.google.com/p/go-uuid/uuid"
	etcderrs "github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors/etcd"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/tools"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/types"

	"github.com/openshift/origin/pkg/user"
	"github.com/openshift/origin/pkg/user/api"
)

// Etcd implements UserIdentityMapping backed by etcd.
type Etcd struct {
	tools.EtcdHelper
	initializer user.Initializer
}

// New returns a new Etcd.
func New(helper tools.EtcdHelper, initializer user.Initializer) *Etcd {
	return &Etcd{
		EtcdHelper:  helper,
		initializer: initializer,
	}
}

var errExists = errors.New("the mapping already exists")

func makeUserKey(id string) string {
	return "/userIdentityMappings/" + id
}

func (r *Etcd) GetUser(name string) (user *api.User, err error) {
	mapping := &api.UserIdentityMapping{}
	err = r.ExtractObj(makeUserKey(name), mapping, false)
	err = etcderrs.InterpretGetError(err, "User", name)
	user = &mapping.User
	return
}

func (r *Etcd) GetUserIdentityMapping(name string) (mapping *api.UserIdentityMapping, err error) {
	mapping = &api.UserIdentityMapping{}
	err = r.ExtractObj(makeUserKey(name), mapping, false)
	err = etcderrs.InterpretGetError(err, "UserIdentityMapping", name)
	return
}

// CreateOrUpdateUserIdentityMapping implements useridentitymapping.Registry
func (r *Etcd) CreateOrUpdateUserIdentityMapping(mapping *api.UserIdentityMapping) (*api.UserIdentityMapping, bool, error) {
	// Create Identity.Name by combining Provider and UserName
	name := fmt.Sprintf("%s:%s", mapping.Identity.Provider, mapping.Identity.UserName)
	mapping.Identity.Name = name
	key := makeUserKey(name)

	// track the object we set into etcd to return
	var found *api.UserIdentityMapping
	var created bool

	err := r.AtomicUpdate(key, &api.UserIdentityMapping{}, func(in runtime.Object) (runtime.Object, error) {
		existing := *in.(*api.UserIdentityMapping)

		// did not previously exist
		if existing.Identity.Name == "" {
			uid := uuid.New()
			existing.User.UID = types.UID(uid)
			existing.User.Name = name
			if err := r.initializer.InitializeUser(&mapping.Identity, &existing.User); err != nil {
				return in, err
			}

			// set these again to prevent bad initialization from messing up data
			existing.User.UID = types.UID(uid)
			existing.User.Name = name
			existing.Identity = mapping.Identity

			found = &existing
			created = true
			return &existing, nil
		}

		if existing.User.Name != name {
			return in, fmt.Errorf("the provided user name does not match the existing mapping %s", existing.User.Name)
		}
		found = &existing

		// TODO: should update identity based on new info as well.
		return in, errExists
	})

	if err != nil && err != errExists {
		err = etcderrs.InterpretCreateError(err, "UserIdentityMapping", name)
		return nil, false, err
	}
	return found, created, nil
}
