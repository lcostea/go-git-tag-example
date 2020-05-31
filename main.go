package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	log "github.com/sirupsen/logrus"
)

func main() {
	//you will probably want to modify this variables, if you are going to use this example
	url := "git@github.com:lcostea/sample-controller-workshop.git"
	dir := "/tmp/sample-controller-workshop"
	name := "Liviu Costea"
	email := "_your_email_address_@gmail.com"
	tag := "v0.1.0"

	r, err := cloneRepo(url, dir)

	if err != nil {
		log.Errorf("clone repo error: %s", err)
		return
	}

	if tagExists(tag, r) {
		log.Infof("Tag %s already exists, nothing to do here", tag)
		return
	}

	created, err := setTag(r, tag, defaultSignature(name, email))
	if err != nil {
		log.Errorf("create tag error: %s", err)
		return
	}

	if created {
		err = pushTags(r)
		if err != nil {
			log.Errorf("push tag error: %s", err)
			return
		}
	}

}

func cloneRepo(url, dir string) (*git.Repository, error) {

	log.Infof("cloning %s into %s", url, dir)
	auth, keyErr := publicKey()
	if keyErr != nil {
		return nil, keyErr
	}

	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		Progress: os.Stdout,
		URL:      url,
		Auth:     auth,
	})

	if err != nil {
		if err == git.ErrRepositoryAlreadyExists {
			log.Info("repo was already cloned")
		} else {
			log.Errorf("clone git repo error: %s", err)
			return nil, err
		}
	}

	return r, nil
}

func publicKey() (*ssh.PublicKeys, error) {
	var publicKey *ssh.PublicKeys
	sshPath := os.Getenv("HOME") + "/.ssh/github_rsa"
	sshKey, _ := ioutil.ReadFile(sshPath)
	publicKey, err := ssh.NewPublicKeys("git", []byte(sshKey), "")
	if err != nil {
		return nil, err
	}
	return publicKey, err
}

func tagExists(tag string, r *git.Repository) bool {
	tagFoundErr := "tag was found"
	tags, err := r.TagObjects()
	if err != nil {
		log.Errorf("get tags error: %s", err)
		return false
	}
	res := false
	err = tags.ForEach(func(t *object.Tag) error {
		if t.Name == tag {
			res = true
			return fmt.Errorf(tagFoundErr)
		}
		return nil
	})
	if err != nil && err.Error() != tagFoundErr {
		log.Errorf("iterate tags error: %s", err)
		return false
	}
	return res
}

func setTag(r *git.Repository, tag string, tagger *object.Signature) (bool, error) {
	if tagExists(tag, r) {
		log.Infof("tag %s already exists", tag)
		return false, nil
	}
	log.Infof("Set tag %s", tag)
	h, err := r.Head()
	if err != nil {
		log.Errorf("get HEAD error: %s", err)
		return false, err
	}

	_, err = r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Tagger:  tagger,
		Message: tag,
	})

	if err != nil {
		log.Errorf("create tag error: %s", err)
		return false, err
	}

	return true, nil
}

func pushTags(r *git.Repository) error {

	auth, _ := publicKey()

	po := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       auth,
	}

	err := r.Push(po)

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			log.Info("origin remote was up to date, no push done")
			return nil
		}
		log.Errorf("push to remote origin error: %s", err)
		return err
	}

	return nil
}

func defaultSignature(name, email string) *object.Signature {
	return &object.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}
}
