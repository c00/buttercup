package daemoncmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type entry struct {
	path     string
	dirEntry fs.DirEntry
}

// Breadth First walk over filsystem, returning files and directories
func BreadthFirstWalk(root string, cb fs.WalkDirFunc) error {
	log.Info("Walking folder: %v", root)

	stat, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("cannot stat root: %w", err)
	}

	//Queue with root entry
	queue := []entry{
		{
			path:     root,
			dirEntry: fs.FileInfoToDirEntry(stat),
		},
	}

	for i := 0; i < len(queue); i++ {
		//current item
		current := queue[i]

		if current.dirEntry.IsDir() {
			entries, err := os.ReadDir(current.path)
			if err != nil {
				return cb(current.path, current.dirEntry, fmt.Errorf("cannot read dir: %w", err))
			}

			for _, item := range entries {
				queue = append(queue, entry{
					path:     filepath.Join(current.path, item.Name()),
					dirEntry: item,
				})
			}
		}

		err = cb(current.path, current.dirEntry, nil)
		if err != nil {
			//Call it again with the error.
			return fmt.Errorf("callback error: %w", err)
		}
	}

	return nil
}

// Breadth First walk over filsystem, returning only the directories
func BreadthFirstDirWalk(root string, cb fs.WalkDirFunc) error {
	return BreadthFirstWalk(root, func(path string, entry fs.DirEntry, err error) error {

		if entry.IsDir() {
			cb(path, entry, err)
		}

		return nil
	})
}
