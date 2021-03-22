package ops

import (
	"crypto/tls"
	"embed"
	"fmt"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"io/fs"
	"log"
	"net/http"
)

//go:embed views
var content embed.FS

func RunServer(address string, port int) (err error) {
	router := gin.Default()

	api := router.Group("/api")

	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}

	api.GET("/systems", SystemHandler)
	//api.POST("/jokes/like/:jokeID", LikeJoke)

	router.Use(Serve("/", content))

	addr := fmt.Sprintf("%s:%d", address, port)
	fmt.Printf("Server starting on %s.\n", addr)

	err = router.Run(addr)

	return err
}

func Serve(urlPrefix string, efs embed.FS) gin.HandlerFunc {
	// the embedded filesystem has a 'views/' at the top level.  We wanna strip this so we can treat the root of the views directory as the web root.
	fsys, err := fs.Sub(efs, "views")
	if err != nil {
		log.Fatalf(err.Error())
	}

	fileserver := http.FileServer(http.FS(fsys))
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}

	return func(c *gin.Context) {
		fileserver.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

func SystemHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	stacks, err := GetStacks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, make([]DisplayStack, 0))
	}

	c.JSON(http.StatusOK, stacks)
}

func GetStacks() (stacks []DisplayStack, err error) {
	config, err := LoadConfig("")
	if err != nil {
		err = errors.Wrapf(err, "failed to load default config file")
		return stacks, err
	}

	s, err := NewStack(config, nil, true)
	if err != nil {
		err = errors.Wrapf(err, "Failed to create devenv object")
		return stacks, err
	}

	stacklist, err := s.ListStacks()
	if err != nil {
		err = errors.Wrapf(err, "Error listing stacks")
		return stacks, err
	}

	output, err := sts.New(s.AwsSession).GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		err = errors.Wrapf(err, "Error getting caller identity")
		return stacks, err
	}

	account := *output.Account

	stacks = make([]DisplayStack, 0)

	for _, stack := range stacklist {
		created := stack.CreationTime.String()
		config.StackName = *stack.StackName
		s, err := NewStack(config, nil, true)
		if err != nil {
			err = errors.Wrapf(err, "Failed to create stack object")
			return stacks, err
		}

		outputs, err := s.Outputs()
		if err != nil {
			err = errors.Wrapf(err, "failed getting outputs for %s", config.StackName)
			return stacks, err
		}

		var address string
		var caHost string
		var api string
		var login string

		for _, o := range outputs {
			switch *o.OutputKey {
			case "Address":
				address = *o.OutputValue

			case "Login":
				e := PingEndpoint(*o.OutputValue)
				if e != nil {
					login = "Not Ready"
				}
				login = fmt.Sprintf("https://%s", *o.OutputValue)

			case "Api":
				e := PingEndpoint(*o.OutputValue)
				if e != nil {
					api = "Not Ready"
				}
				api = fmt.Sprintf("https://%s", *o.OutputValue)

			case "CA":
				e := PingEndpoint(*o.OutputValue)
				if e != nil {
					api = "Not Ready"
				}
				caHost = fmt.Sprintf("https://%s/v1/pki/ca/pem", *o.OutputValue)
			}
		}

		display := DisplayStack{
			Account:  account,
			Name:     *stack.StackName,
			CFStatus: *stack.StackStatus,
			Address:  address,
			Kotsadm:  fmt.Sprintf("http://%s:8800", address),
			Api:      api,
			Login:    login,
			CA:       caHost,
			Created:  created,
		}

		stacks = append(stacks, display)
	}

	return stacks, err
}

type DisplayStack struct {
	Account     string `json:"account" binding:"required"`
	Kubernetes  string `json:"kubernetes" binding:"required"`
	Kotsadm     string `json:"kotsadm" binding:"required"`
	CFStatus    string `json:"cfstatus" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Address     string `json:"address" binding:"required"`
	Datastore   string `json:"datastore" binding:"required"`
	EventStream string `json:"eventstream" binding:"required"`
	Media       string `json:"media" binding:"required"`
	Login       string `json:"login" binding:"required"`
	Api         string `json:"api" binding:"required"`
	CDN         string `json:"cdn" binding:"required"`
	CA          string `json:"ca" binding:"required"`
	Created     string `json:"created" binding:"required"`
}

func PingEndpoint(address string) (err error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	_, err = http.Get(fmt.Sprintf("https://%s", address))

	return err
}

//func LikeJoke(c *gin.Context) {
//	if jokeid, err := strconv.Atoi(c.Param("jokeID")); err == nil {
//		for i:=0; i < len(jokes); i++ {
//			if jokes[i].ID == jokeid {
//				jokes[i].Likes += 1
//			}
//		}
//
//		c.JSON(http.StatusOK, &jokes)
//	} else {
//		c.AbortWithStatus(http.StatusNotFound)
//	}
//}

// stack name , account, statuses,
