package fs

import (
	"io"
	"os"

	"github.com/Microsoft/go-winio"
	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

const (
	seTakeOwnershipPrivilege = "SeTakeOwnershipPrivilege"
)

func (c *copier) copyFileInfo(fi os.FileInfo, src, name string) error {
	if err := os.Chmod(name, fi.Mode()); err != nil {
		return errors.Wrapf(err, "failed to chmod %s", name)
	}

	// Copy file ownership and ACL
	// We need SeRestorePrivilege and SeTakeOwnershipPrivilege in order
	// to restore security info on a file, especially if we're trying to
	// apply security info which includes SIDs not necessarily present on
	// the host.
	privileges := []string{winio.SeRestorePrivilege, seTakeOwnershipPrivilege}
	if err := winio.EnableProcessPrivileges(privileges); err != nil {
		return err
	}
	defer winio.DisableProcessPrivileges(privileges)

	secInfo, err := windows.GetNamedSecurityInfo(
		src, windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)

	if err != nil {
		return err
	}

	dacl, _, err := secInfo.DACL()
	if err != nil {
		return err
	}

	sid, _, err := secInfo.Owner()
	if err != nil {
		return err
	}

	if err := windows.SetNamedSecurityInfo(
		name, windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION,
		sid, nil, dacl, nil); err != nil {

		return err
	}
	return nil
}

func (c *copier) copyFileTimestamp(fi os.FileInfo, name string) error {
	// TODO: copy windows specific metadata

	return nil
}

func copyFile(source, target string) error {
	src, err := os.Open(source)
	if err != nil {
		return errors.Wrapf(err, "failed to open source %s", source)
	}
	defer src.Close()
	tgt, err := os.Create(target)
	if err != nil {
		return errors.Wrapf(err, "failed to open target %s", target)
	}
	defer tgt.Close()

	return copyFileContent(tgt, src)
}

func copyFileContent(dst, src *os.File) error {
	buf := bufferPool.Get().(*[]byte)
	_, err := io.CopyBuffer(dst, src, *buf)
	bufferPool.Put(buf)
	return err
}

func copyXAttrs(dst, src string, xeh XAttrErrorHandler) error {
	return nil
}

func copyDevice(dst string, fi os.FileInfo) error {
	return errors.New("device copy not supported")
}
