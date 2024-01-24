package benthosldap

import (
	"context"
	"encoding/json"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/go-ldap/ldap/v3"
)

func ldapConfigSpec() *service.ConfigSpec {
	return service.NewConfigSpec().
		// Stable(). TODO
		Version("0.0.1").
		Categories("Services").
		Summary("").
		Description(``).
		Field(service.NewStringField("url").Description("LDAP URL.")).
		Field(service.NewStringField("username").Description("LDAP username.")).
		Field(service.NewStringField("password").Description("LDAP password.")).
		Field(service.NewIntField("buffer_size").Default(2000).Description("LDAP buffer size.")).
		Field(service.NewStringField("filter").Description("LDAP search filter.")).
		Field(service.NewStringField("base").Description("LDAP search base.")).
		Field(service.NewStringListField("attributes").Description("LDAP attributes"))
}

func init() {
	err := service.RegisterInput(
		"ldap", ldapConfigSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Input, error) {
			return newLdapInput(conf)
		})
	if err != nil {
		panic(err)
	}
}

func newLdapInput(conf *service.ParsedConfig) (service.Input, error) {
	url, err := conf.FieldString("url")
	if err != nil {
		return nil, err
	}
	username, err := conf.FieldString("username")
	if err != nil {
		return nil, err
	}
	password, err := conf.FieldString("password")
	if err != nil {
		return nil, err
	}
	bufferSize, err := conf.FieldInt("buffer_size")
	if err != nil {
		return nil, err
	}
	filter, err := conf.FieldString("filter")
	if err != nil {
		return nil, err
	}
	base, err := conf.FieldString("base")
	if err != nil {
		return nil, err
	}
	attrs, err := conf.FieldStringList("attributes")
	if err != nil {
		return nil, err
	}

	conn, err := ldap.DialURL(url)
	if err != nil {
		return nil, err
	}

	if err = conn.Bind(username, password); err != nil {
		return nil, err
	}

	return service.AutoRetryNacks(&ldapInput{
		conn:       conn,
		bufferSize: bufferSize,
		filter:     filter,
		base:       base,
		attrs:      attrs,
	}), nil
}

type ldapInput struct {
	conn       *ldap.Conn
	bufferSize int
	filter     string
	base       string
	attrs      []string
	res        ldap.Response
}

func (l *ldapInput) Connect(ctx context.Context) error {
	if l.res != nil {
		return nil
	}

	req := ldap.NewSearchRequest(
		l.base,
		ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0, 0, false,
		l.filter,
		l.attrs,
		[]ldap.Control{},
	)

	res := l.conn.SearchAsync(ctx, req, l.bufferSize)
	if err := res.Err(); err != nil {
		return err
	}

	l.res = res

	return nil
}

func (l *ldapInput) Read(ctx context.Context) (*service.Message, service.AckFunc, error) {
	if l.conn == nil {
		return nil, nil, service.ErrNotConnected
	}

	next := l.res.Next()

	if err := l.res.Err(); err != nil {
		return nil, nil, err
	}

	if !next {
		return nil, nil, service.ErrEndOfInput
	}

	entry := l.res.Entry()
	vals := make(map[string][]string, len(entry.Attributes))
	for _, a := range entry.Attributes {
		vals[a.Name] = a.Values
	}

	data, err := json.Marshal(vals)
	if err != nil {
		return nil, nil, err
	}

	msg := service.NewMessage(nil)
	msg.SetBytes(data)
	return msg, func(ctx context.Context, err error) error {
		return nil
	}, nil
}

func (l *ldapInput) Close(ctx context.Context) error {
	if l.conn != nil {
		return l.conn.Close()
	}
	return nil
}
