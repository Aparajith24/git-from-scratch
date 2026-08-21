package objects

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Signature is the "<name> <email> <unix-seconds> <+/-HHMM>" that appears on
// both the author and committer lines. Git's format has no matching
// time.Format layout, so the offset is built by hand from When.Zone().
type Signature struct {
	Name  string
	Email string
	When  time.Time
}

func (s Signature) String() string {
	_, offsetSeconds := s.When.Zone() // returns timezone and offset from utc in seconds
	sign := "+"                       //ahead of utc
	if offsetSeconds < 0 {            //behind utc
		sign = "-"
		offsetSeconds = -offsetSeconds
	}
	hours := offsetSeconds / 3600
	minutes := (offsetSeconds % 3600) / 60
	return fmt.Sprintf("%s <%s> %d %s%02d%02d", s.Name, s.Email, s.When.Unix(), sign, hours, minutes)
}

// ParseSignature reverses Signature.String().
func ParseSignature(line string) (Signature, error) {
	// "<name> <email> <unix> <+/-HHMM>" — name may contain spaces, so split
	// from the right: offset, then unix seconds, then "<email>", then name.
	fields := strings.Fields(line) //get the fields from the line string
	if len(fields) < 4 {
		return Signature{}, fmt.Errorf("invalid signature %q", line)
	}
	offsetStr := fields[len(fields)-1]                //offset from utc
	unixStr := fields[len(fields)-2]                  // the unix seconds
	email := fields[len(fields)-3]                    //email
	name := strings.Join(fields[:len(fields)-3], " ") //name

	if !strings.HasPrefix(email, "<") || !strings.HasSuffix(email, ">") {
		return Signature{}, fmt.Errorf("invalid signature %q: bad email field %q", line, email)
	}
	email = strings.TrimSuffix(strings.TrimPrefix(email, "<"), ">")

	unixSeconds, err := strconv.ParseInt(unixStr, 10, 64) //convert unix to string
	if err != nil {
		return Signature{}, fmt.Errorf("invalid signature %q: bad timestamp: %w", line, err)
	}

	if len(offsetStr) != 5 || (offsetStr[0] != '+' && offsetStr[0] != '-') {
		return Signature{}, fmt.Errorf("invalid signature %q: bad offset %q", line, offsetStr)
	}
	offsetHours, err1 := strconv.Atoi(offsetStr[1:3])
	offsetMinutes, err2 := strconv.Atoi(offsetStr[3:5])
	if err1 != nil || err2 != nil {
		return Signature{}, fmt.Errorf("invalid signature %q: bad offset %q", line, offsetStr)
	}
	offsetSeconds := offsetHours*3600 + offsetMinutes*60
	if offsetStr[0] == '-' {
		offsetSeconds = -offsetSeconds
	}

	loc := time.FixedZone(offsetStr, offsetSeconds)
	return Signature{Name: name, Email: email, When: time.Unix(unixSeconds, 0).In(loc)}, nil
}

// Commit points at a tree plus zero or more parent commits — it's the
// object that turns content-addressed storage into history.
type Commit struct {
	Tree      string
	Parents   []string // empty for a root commit, 2+ for a merge
	Author    Signature
	Committer Signature
	Message   string
}

func (c Commit) Type() string {
	return "commit"
}

// Serialize encodes the commit in git's exact field order: tree, one parent
// line per parent, author, committer, a blank line, then the message.
func (c Commit) Serialize() []byte {
	var buf bytes.Buffer //initialize a buffer
	fmt.Fprintf(&buf, "tree %s\n", c.Tree)
	for _, parent := range c.Parents {
		fmt.Fprintf(&buf, "parent %s\n", parent)
	}
	fmt.Fprintf(&buf, "author %s\n", c.Author)
	fmt.Fprintf(&buf, "committer %s\n", c.Committer)
	buf.WriteByte('\n')
	buf.WriteString(c.Message)
	if !strings.HasSuffix(c.Message, "\n") {
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

// Save serializes and writes the commit to the object store.
func (c Commit) Save(gitDir string) (string, error) {
	return WriteObject(gitDir, "commit", c.Serialize())
}

// ParseCommit reverses Serialize: given a commit object's raw data (as
// returned by ReadObject), it recovers the struct.
func ParseCommit(data []byte) (Commit, error) {
	r := bufio.NewScanner(bytes.NewReader(data))
	r.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	c := Commit{}
	for r.Scan() {
		line := r.Text()
		if line == "" {
			break // blank line separates headers from the message
		}

		key, value, ok := strings.Cut(line, " ")
		if !ok {
			return Commit{}, fmt.Errorf("invalid commit header line %q", line)
		}

		switch key {
		case "tree":
			c.Tree = value
		case "parent":
			c.Parents = append(c.Parents, value)
		case "author":
			sig, err := ParseSignature(value)
			if err != nil {
				return Commit{}, err
			}
			c.Author = sig
		case "committer":
			sig, err := ParseSignature(value)
			if err != nil {
				return Commit{}, err
			}
			c.Committer = sig
		default:
			return Commit{}, fmt.Errorf("unrecognized commit header %q", key)
		}
	}
	if err := r.Err(); err != nil {
		return Commit{}, err
	}

	var message bytes.Buffer
	for r.Scan() {
		message.WriteString(r.Text())
		message.WriteByte('\n')
	}
	if err := r.Err(); err != nil {
		return Commit{}, err
	}
	c.Message = message.String()

	if c.Tree == "" {
		return Commit{}, fmt.Errorf("commit missing tree header")
	}
	return c, nil
}
