package lock

import (
	"crypto/md5"
	"fmt"
	"sync"

	"github.com/grupawp/tensorflow-deploy/app"
	"github.com/grupawp/tensorflow-deploy/exterr"
)

const (
	maxKeyLength      = app.MaxTeamLength + app.MaxProjectLength + app.MaxModelNameLength
	errWrongKeyLength = "wrong key length"
)

var (
	lockWrongKeyLengthErrorCode = 1001
	lockKeyIsLockedErrorCode    = 1002

	errKeyIsLocked = exterr.NewErrorWithMessage("the key is already locked").WithComponent(app.ComponentLock).WithCode(lockKeyIsLockedErrorCode)
)

type Lock struct {
	state sync.Map
}

// New - create lock object
func New() *Lock {
	return &Lock{}
}

func (l *Lock) lock(key string) error {
	md5Key := md5.Sum([]byte(key))
	_, loaded := l.state.LoadOrStore(md5Key, 0)
	if loaded {
		return errKeyIsLocked
	}
	return nil
}

func (l *Lock) unLock(key string) {
	md5Key := md5.Sum([]byte(key))
	l.state.Delete(md5Key)
}

// Lock - setting lock on team/project/name item or returns error if item is already locked
func (l *Lock) Lock(servable app.ServableID) error {
	key := servable.Team + servable.Project + servable.Name
	if len(key) == 0 || len(key) > maxKeyLength {
		return exterr.NewErrorWithMessage(fmt.Sprintf("%s, max length: %d", errWrongKeyLength, maxKeyLength)).
			WithComponent(app.ComponentLock).WithCode(lockWrongKeyLengthErrorCode)
	}
	return l.lock(key)
}

// UnLock - remove lock from team/project/name item
func (l *Lock) UnLock(servable app.ServableID) {
	key := servable.Team + servable.Project + servable.Name
	l.unLock(key)
}

func (l *Lock) LockID(id string) error {
	return l.lock(id)
}

func (l *Lock) UnLockID(id string) {
	l.unLock(id)
}

func (l *Lock) IsLockedID(keyID string) bool {
	md5Key := md5.Sum([]byte(keyID))
	result := false
	l.state.Range(func(key interface{}, value interface{}) bool {
		if key == md5Key {
			result = true
			return false
		}
		return true
	})

	return result
}
