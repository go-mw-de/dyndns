//AWS DNS Update part of this Programm is based on https://github.com/agorf/dyndns53

package dyndns

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"gitlab.com/echtwerner/appengine/auth"
	"gitlab.com/echtwerner/appengine/collection"
	"gitlab.com/echtwerner/appengine/reqhelper"
)

const (
	USAGE = "Parameters: ipv4, ipv6(not implemented), hostname"
)

type Handler struct {
	Collection collection.Collection
	Realm      string
}

func NewHandler(c collection.Collection, realm string) Handler {
	return Handler{
		Collection: c,
		Realm:      realm,
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get and prepare Contaxt
	ctx := r.Context()
	ctx = context.WithValue(ctx, "Realm", h.Realm)
	// Set  Request ID for http.Request. This will be used to create a logger with fields
	r = r.WithContext(reqhelper.AssignLogger(ctx))
	// Setup logger with RequestID
	reqID := reqhelper.GetRequestID(ctx)
	logger := log.WithField("LogFieldKeyRequestID", reqID)
	//Get HandlerFunc
	collection := h.Collection
	hFunc := dyndns(ctx, collection)

	logger.Infof("Incomming request %s %s %s. Authorization Required", r.Method, r.RequestURI, r.RemoteAddr)
	// Enable Authorization
	hFunc = auth.BasicAuthHandler(ctx, hFunc, collection.GetUser, h.Realm)
	hFunc(w, r)
	logger.Infof("Finished handling http req.")
}

type Domain interface {
	Add(hostname string, ipv4 string) error
	Delete(hostname string, ipv4 string) error
	Update(hostname string, ipv4 string) error
}

var ipAddresses map[string]string

// dyndns will return a regular http.HandlerFunc but it is necessary to use my own datastructes in this handlerfFunc.
func dyndns(ctx context.Context, collection collection.Collection) http.HandlerFunc {
	// Setup logger with RequestID
	reqID := reqhelper.GetRequestID(ctx)
	logger := log.WithField("LogFieldKeyRequestID", reqID)
	logger.Infof("dyndns(): outter called")
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqID := reqhelper.GetRequestID(ctx)
		// let's use it with logrus FieldLogger!
		logger := log.WithField("LogFieldKeyRequestID", reqID)
		ctxUser := ctx.Value("User")
		ctxRealm := ctx.Value("App")
		logger.Infof("dyndns.HandleFunc(): called for user '%s' with realm '%s'", ctxUser, ctxRealm)

		// Process http.Request
		var hname, ipv4 string
		var quiet bool
		switch r.Method {
		case "GET":
			logger.Infof("GET Request: Url Params: '%v'", r.URL)
			myips, ok := r.URL.Query()["ipv4"]
			if !ok || len(myips[0]) < 1 {
				ipv4 = getIP(r)
				logger.Infof("Url Param 'ipv4' is missing, used Client IP %s instead", ipv4)
			} else {
				// Query()["key"] will return an array of items,
				// we only want the single item.
				ipv4 = myips[0]
			}
			hnames, ok := r.URL.Query()["hostname"]
			if !ok || len(hnames[0]) < 1 {
				fmt.Fprintln(w, USAGE)
				logger.Infof("Url Param 'hostname' is missing")
				return
			}
			hname = hnames[0]
			// Print Debug Messages
			logger.Infof("GET Request: Url Params: ipv4: %s, hostname: %s", ipv4, hname)

			// For Firtz.box there is a workaround where we do not want to send an answer.
			// Just check if the parameter is set.
			_, quiet = r.URL.Query()["quiet"]
		default:
			logger.Infof("Only GET methods are supported.")
			return
		}

		user, _ := collection.GetUser(ctx, ctxUser.(string))
		logger.Infof("dyndns(): login: '%s' domains:  '%v'", user.Login, user.Domains)

		if !user.VerifyDomain(hname) {
			fmt.Fprintf(w, "Sorry, you '%s' are not allowed to update '%s'.\n", user.Login, hname)
			logger.Infof("User '%s' is not allowed to update host'%s'. Allowed host'%#v'.", user.Login, hname, user.Domains)
			return
		}

		nameParts := strings.Split(hname, ".")
		// expect the firstname in NameParts are the subdomain
		d := getDomain(ctx, collection, strings.Join(nameParts[1:], "."))

		if err := d.Update(hname, ipv4); err != nil {
			logger.Fatalf("dyndns(): Something went wrong: %v", err)
		}
		logger.Infof("dydns(): current IP address for %s is %s; upsert request sent", hname, ipv4)
		if quiet {
			logger.Info("Be quiet - No output requested")
			return
		}
		output := fmt.Sprintf("DYDNS: hostname %s updated with ip %s\n", hname, ipv4)
		fmt.Fprintf(w, output)
	}
}

func getDomain(ctx context.Context, c collection.Collection, k string) (rval Domain) {
	raw, _ := c.GetDomain(ctx, k)
	switch dp := raw.Dnsprovider; dp {
	case "aws":
		rval = DomainAWS{Data: raw}
	case "gae":
		rval = DomainGAE{Data: raw}
	}
	return rval
}

func getIP(r *http.Request) string {
	if ipProxy := r.Header.Get("X-Forwarded-For"); len(ipProxy) > 0 {
		return strings.Split(ipProxy, ",")[0]
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
