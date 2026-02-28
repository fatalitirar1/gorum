package gormal

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	// (c_iflag)
	IGNBRK IFlag = unix.IGNBRK
	BRKINT IFlag = unix.BRKINT
	IGNPAR IFlag = unix.IGNPAR
	PARMRK IFlag = unix.PARMRK
	INPCK  IFlag = unix.INPCK
	ISTRIP IFlag = unix.ISTRIP
	INLCR  IFlag = unix.INLCR
	IGNCR  IFlag = unix.IGNCR
	ICRNL  IFlag = unix.ICRNL
	IXON   IFlag = unix.IXON
	IXOFF  IFlag = unix.IXOFF

	// (c_oflag)
	OPOST OFlag = unix.OPOST
	ONLCR OFlag = unix.ONLCR

	// (c_cflag)
	CSIZE  CFlag = unix.CSIZE
	CS5    CFlag = unix.CS5
	CS6    CFlag = unix.CS6
	CS7    CFlag = unix.CS7
	CS8    CFlag = unix.CS8
	CSTOPB CFlag = unix.CSTOPB
	CREAD  CFlag = unix.CREAD
	PARENB CFlag = unix.PARENB
	PARODD CFlag = unix.PARODD
	HUPCL  CFlag = unix.HUPCL
	CLOCAL CFlag = unix.CLOCAL

	// (c_lflag)
	ISIG   LFlag = unix.ISIG
	ICANON LFlag = unix.ICANON
	ECHO   LFlag = unix.ECHO
	ECHOE  LFlag = unix.ECHOE
	ECHOK  LFlag = unix.ECHOK
	ECHONL LFlag = unix.ECHONL
	NOFLSH LFlag = unix.NOFLSH
	TOSTOP LFlag = unix.TOSTOP
	IEXTEN LFlag = unix.IEXTEN
)

type IFlag uint32

type OFlag uint32

type CFlag uint32

type LFlag uint32

type Flag interface {
	GetFlagName() string
	uint32() uint32
}

type Gormal struct {
	fd           uintptr
	termiosFlags map[string]uint32
	bTermios     map[string]uint32
}

func (i IFlag) GetFlagName() string {
	return "Iflag"
}

func (i IFlag) uint32() uint32 {
	return uint32(i)
}

func (o OFlag) GetFlagName() string {
	return "Oflag"
}

func (o OFlag) uint32() uint32 {
	return uint32(o)
}
func (c CFlag) GetFlagName() string {
	return "Cflag"
}

func (c CFlag) uint32() uint32 {
	return uint32(c)
}

func (l LFlag) GetFlagName() string {
	return "Lflag"
}
func (l LFlag) uint32() uint32 {
	return uint32(l)
}

func NewGormalStdin() (*Gormal, error) {
	return NewGormalFromDesctiptor(os.Stdin.Fd())
}
func NewGormalStdOut() (*Gormal, error) {
	return NewGormalFromDesctiptor(os.Stdout.Fd())
}
func NewGormalStdErr() (*Gormal, error) {
	return NewGormalFromDesctiptor(os.Stdout.Fd())
}

func NewGormalFromDesctiptor(fd uintptr) (*Gormal, error) {
	gorm := &Gormal{}
	gorm.termiosFlags = make(map[string]uint32)

	err := gorm.tCGet(fd)
	if err != nil {
		return nil, err
	}

	backup := make(map[string]uint32)
	for k, v := range gorm.termiosFlags {
		backup[k] = v
	}
	gorm.bTermios = backup

	gorm.fd = fd

	return gorm, nil
}

func (gorm *Gormal) AppendFlag(flag Flag) error {
	section := flag.GetFlagName()
	return gorm.AppendFlagToSection(flag, section)
}

func (gorm *Gormal) AppendFlagToSection(flag Flag, section string) error {
	uFlag := flag.uint32()
	return gorm.RowAppendFlagToSection(uFlag, section)
}

func (gorm *Gormal) RowAppendFlagToSection(uFlag uint32, section string) error {

	_, ok := gorm.termiosFlags[section]
	if !ok {
		return fmt.Errorf("section not found")
	}

	if gorm.termiosFlags[section]&uFlag != 0 {
		return fmt.Errorf("flag %#x already set", uFlag)
	}
	gorm.termiosFlags[section] |= uFlag

	return gorm.tCSet(gorm.fd)
}

func (gorm *Gormal) DropFlag(flag Flag) error {
	section := flag.GetFlagName()
	return gorm.DropFlagFromSection(flag, section)
}

func (gorm *Gormal) DropFlagFromSection(flag Flag, section string) error {
	uFlag := flag.uint32()
	return gorm.RowDropFlagFromSection(uFlag, section)
}

func (gorm *Gormal) RowDropFlagFromSection(uFlag uint32, section string) error {

	_, ok := gorm.termiosFlags[section]
	if !ok {
		return fmt.Errorf("section not found")
	}
	gorm.termiosFlags[section] &^= uFlag
	return gorm.tCSet(gorm.fd)
}

func (gorm *Gormal) CheckFlag(flag Flag) (bool, error) {
	section := flag.GetFlagName()
	return gorm.CheckFlagInSection(flag, section)
}

func (gorm *Gormal) CheckFlagInSection(flag Flag, section string) (bool, error) {
	uFlag := flag.uint32()
	return gorm.CheckRowFlaginSection(uFlag, section)
}

func (gorm *Gormal) CheckRowFlaginSection(uFlag uint32, section string) (bool, error) {

	_, ok := gorm.termiosFlags[section]
	if !ok {
		return false, fmt.Errorf("section not found")

	}

	return gorm.termiosFlags[section]&uFlag > 0, nil
}

func (gorm *Gormal) GetTermios() *unix.Termios {
	t := &unix.Termios{}
	gorm.MapToTermios(t)
	return t
}

func (gorm *Gormal) Restore() error {
	flagMap := make(map[string]uint32)
	for k, v := range gorm.termiosFlags {
		flagMap[k] = v
	}
	gorm.termiosFlags = flagMap
	return gorm.tCSet(gorm.fd)
}

func (gorm *Gormal) tCGet(fd uintptr) error {
	t := &unix.Termios{}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		unix.TCGETS,
		uintptr(unsafe.Pointer(t)),
	)
	if errno != 0 {
		return fmt.Errorf("gormal tcGET [ %v ] ", errno)
	}

	gorm.TemiosToMap(t)
	return nil
}

func (gorm *Gormal) tCSet(fd uintptr) error {

	t := &unix.Termios{}
	gorm.MapToTermios(t)

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		unix.TCSETS,
		uintptr(unsafe.Pointer(t)),
	)
	if errno != 0 {
		return fmt.Errorf("gormal TC_SET [ %v ] ", errno)
	}

	gorm.TemiosToMap(t)
	return nil
}

func (gorm *Gormal) TemiosToMap(t *unix.Termios) {
	gorm.termiosFlags["Iflag"] = t.Iflag
	gorm.termiosFlags["Oflag"] = t.Oflag
	gorm.termiosFlags["Cflag"] = t.Cflag
	gorm.termiosFlags["Lflag"] = t.Lflag
}

func (gorm *Gormal) MapToTermios(t *unix.Termios) {
	t.Iflag = gorm.termiosFlags["Iflag"]
	t.Oflag = gorm.termiosFlags["Oflag"]
	t.Cflag = gorm.termiosFlags["Cflag"]
	t.Lflag = gorm.termiosFlags["Lflag"]
}
