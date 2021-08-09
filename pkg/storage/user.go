package storage

import (
	"fmt"

	"github.com/t-bfame/diago/pkg/model"
	"github.com/t-bfame/diago/pkg/tools"

	"github.com/boltdb/bolt"
	log "github.com/sirupsen/logrus"
)

// This is the boltDB bucket name for storing "model/User".
const UserBucketName = "User"

// Initialize botDB for "model/User" storage.
func initStorageUser(db *bolt.DB) error {
	if err := db.Update(createInitBucketFunc(UserBucketName)); err != nil {
		return err
	}
	return nil
}

// Add a "model/User" to the storage.
func AddUser(user *model.User) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UserBucketName))
		if b == nil {
			return fmt.Errorf("missing bucket '%s'", UserBucketName)
		}
		enc, err := tools.GobEncode(user)
		if err != nil {
			return fmt.Errorf("failed to encode User due to: %s", err)
		}
		if err := b.Put([]byte(user.ID), enc); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("user", user).Error("Failed to add User")
		return err
	}
	return nil
}

// Delete a "model/User" with the specified UserID from the storage.
func DeleteUser(userID model.UserID) error {
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UserBucketName))
		if b == nil {
			return fmt.Errorf("missing bucket '%s'", UserBucketName)
		}
		if err := b.Delete([]byte(userID)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("userID", userID).Error("Failed to delete User")
		return err
	}
	return nil
}

// Retrieve a "model/User" with the specified UserID from the storage.
func GetUserByUserId(userId model.UserID) (*model.User, error) {
	var result *model.User
	if err := db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(UserBucketName))
		data := b.Get([]byte(userId))
		if data == nil {
			return nil
		}
		if err = tools.GobDecode(&result, data); err != nil {
			return fmt.Errorf("failed to decode User due to: %s", err)
		}
		return nil
	}); err != nil {
		log.WithError(err).WithField("userId", userId).Error("Failed to GetUserById")
	}
	return result, nil
}

// Retrieve all "model/User" stored in the storage.
func GetAllUsers() ([]*model.User, error) {
	var users = make([]*model.User, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(UserBucketName))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var user *model.User
			if err := tools.GobDecode(&user, v); err != nil {
				return fmt.Errorf("failed to decode User due to: %s", err)
			}
			users = append(users, user)
		}

		return nil
	}); err != nil {
		log.WithError(err).Error("Failed to GetAllUsers")
		return nil, err
	}
	return users, nil
}
