package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/didier13150/gitlablib"
)

type arrayFlags []string

type GLPipelineVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GLPipelineData struct {
	Ref       string          `json:"ref"`
	Variables []GLPipelineVar `json:"variables"`
}

type GLPipelineReturnDetailedStatus struct {
	Id           int    `json:"id"`
	Icon         string `json:"icon"`
	Text         string `json:"text"`
	Label        string `json:"label"`
	Group        string `json:"group"`
	Tooltip      string `json:"tooltip"`
	HasDetails   bool   `json:"has_details"`
	DetailsPath  string `json:"details_path"`
	Illustration string `json:"illustration"`
	Favicon      string `json:"favicon"`
}

type GLPipelineReturnUser struct {
	Id          int    `json:"id"`
	Username    string `json:"username"`
	PublicEmail string `json:"public_email"`
	Name        string `json:"name"`
	State       string `json:"state"`
	Locked      bool   `json:"locked"`
	AvatarUrl   string `json:"avatar_url"`
	WebUrl      string `json:"web_url"`
}

type GLPipelineReturn struct {
	Id             int                            `json:"id"`
	IId            int                            `json:"iid"`
	ProjectId      int                            `json:"project_id"`
	Sha            string                         `json:"sha"`
	Ref            string                         `json:"ref"`
	Status         string                         `json:"status"`
	Source         string                         `json:"source"`
	CreatedAt      string                         `json:"created_at"`
	UpdatedAt      string                         `json:"updated_at"`
	WebUrl         string                         `json:"web_url"`
	BeforeSha      string                         `json:"before_sha"`
	Tag            bool                           `json:"tag"`
	YamlError      bool                           `json:"yaml_errors"`
	User           GLPipelineReturnUser           `json:"user"`
	StartedAt      string                         `json:"started_at"`
	FinishedAt     string                         `json:"finished_at"`
	CommitedAt     string                         `json:"committed_at"`
	Duration       int                            `json:"duration"`
	QueueDuration  int                            `json:"queued_duration"`
	Coverage       string                         `json:"coverage"`
	DetailedStatus GLPipelineReturnDetailedStatus `json:"detailed_status"`
	Archived       bool                           `json:"archived"`
}

// String is an implementation of the flag.Value interface
func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

// Set is an implementation of the flag.Value interface
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func getDefaultValue(envVar string, defaultVar string) string {

	if len(os.Getenv(envVar)) > 0 {
		return os.Getenv(envVar)
	}
	return defaultVar
}

func main() {

	var projects gitlablib.GitlabProject
	var varList arrayFlags

	flag.Var(&varList, "var", "Var for pipeline, this option can be specified more than one time.")
	var projectId = flag.String("id", "", "Gitlab project identifiant.")
	var projectIdFile = flag.String("idfile", getDefaultValue("GLCLI_ID_FILE", ".gitlab.id"), "Gitlab project identifiant file.")
	var projectsFile = flag.String("projectfile", getDefaultValue("GLCLI_PROJECT_FILE", os.Getenv("HOME")+"/.gitlab-projects.json"), "File which contains projects.")
	var gitlabUrl = flag.String("url", getDefaultValue("GLCLI_GITLAB_URL", "https://gitlab.com"), "Gitlab URL.")
	var gitlabTokenFile = flag.String("tokenfile", getDefaultValue("GLCLI_TOKEN_FILE", os.Getenv("HOME")+"/.gitlab.token"), "File which contains token to access Gitlab API.")
	var remoteName = flag.String("remote", getDefaultValue("GLCLI_REMOTE_NAME", "origin"), "Git remote name.")
	var branchName = flag.String("branch", "", "Git branch.")
	var verboseMode = flag.Bool("verbose", false, "Make application more talkative.")
	var debugMode = flag.Bool("debug", false, "Enable debug mode")
	var dryrunMode = flag.Bool("dryrun", false, "Run in dry-run mode (read only).")
	flag.Usage = func() {
		fmt.Print("Run pipeline for current project\n\n")
		fmt.Printf("Usage: " + os.Args[0] + " [options]\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *verboseMode {
		log.Print("Verbose mode is active")
	}
	if *debugMode {
		log.Print("Debug mode is active")
	}
	if *dryrunMode {
		log.Print("Dry-run mode is active")
	}
	if projectIdFile != nil {
		log.Printf("Project id file: %s", *projectIdFile)
	}
	if projectsFile != nil {
		log.Printf("Projects file: %s", *projectsFile)
	}
	if gitlabUrl != nil {
		log.Printf("Gitlab URL: %s", *gitlabUrl)
	}
	if gitlabTokenFile != nil {
		log.Printf("Gitlab Token file: %s", *gitlabTokenFile)
	}
	if projectId != nil {
		log.Printf("Project ID: %s", *projectId)
	}
	if remoteName != nil {
		log.Printf("Git remote name: %s", *remoteName)
	}
	if branchName != nil {
		log.Printf("Git branch name: %s", *branchName)
	}

	for _, envVar := range varList {
		log.Printf("Var: %s", envVar)
	}

	token := gitlablib.ReadFromFile(*gitlabTokenFile, "token", *verboseMode)
	log.Printf("Token: %s", token)

	if *branchName == "" {
		if *projectId == "" {
			head := gitlablib.ReadFromFile(".git/HEAD", "git head", *verboseMode)
			ref := strings.Split(head, "/")[len(strings.Split(head, "/"))-1]
			*branchName = ref
			log.Printf("Current branch is %s", *branchName)
		} else {
			log.Fatal("Branch option must be specified when using id option.")
		}
	}

	projectfile, err := os.OpenFile(*projectsFile, os.O_RDONLY, 0644)
	if err == nil {
		projects.ImportProjects(*projectsFile)
		err = projectfile.Close()
		if err != nil {
			log.Fatalln("Cannot close project file")
		}
		if *projectId == "" {
			repoUrl := gitlablib.GetGitUrl(*remoteName, *verboseMode)
			if *verboseMode {
				log.Printf("Get git repository url for remote %s: %s", *remoteName, repoUrl)
			}
			id := projects.GetProjectIdByRepoUrl(repoUrl)
			if id > 0 {
				*projectId = strconv.Itoa(id)
				if *verboseMode {
					log.Printf("Get projectId: %s from git repository URL %s", *projectId, repoUrl)
				}
			}
		}
	} else {
		if *verboseMode {
			log.Printf("Cannot open %s file", *projectsFile)
		}
	}

	var data GLPipelineData
	data.Ref = *branchName

	for _, envVar := range varList {
		keyval := strings.SplitN(envVar, "=", 2)
		var pvardata GLPipelineVar
		pvardata.Key = keyval[0]
		pvardata.Value = keyval[1]
		data.Variables = append(data.Variables, pvardata)
	}

	datajson, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
		return
	}

	uri := fmt.Sprintf("/api/v4/projects/%s/pipeline", *projectId)
	if *verboseMode {
		log.Printf("URI: %s", uri)
	}

	if *dryrunMode || *verboseMode {
		fmt.Printf("Url: %s\n", *gitlabUrl+uri)
		fmt.Printf("Data: %s\n", datajson)
		if *dryrunMode {
			return
		}
	}
	glApi := gitlablib.NewGLApi(*gitlabUrl, token, *verboseMode)
	retjson, _, err := glApi.CallGitlabApi(uri, "POST", datajson)
	if err != nil {
		log.Println(err)
		return
	}
	var ret GLPipelineReturn
	err = json.Unmarshal(retjson, &ret)
	if err != nil {
		log.Println(err)
		fmt.Print(string(retjson))
		return
	}
	log.Printf("Pipeline %s: %s", ret.Status, ret.WebUrl)
	if *verboseMode {
		fmt.Print(string(retjson))
	}
}
