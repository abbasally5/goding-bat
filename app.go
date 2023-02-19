package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	//"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	HomeColumns = 3
	ProbColumns = 3
)

var baseTemplates = []string{
	"templates/layout/footer.tmpl",
	"templates/layout/header.tmpl",
	"templates/layout/base.tmpl",
}

// Models
type ProblemSetOverview struct {
	Name        string `bson:"problem_set_name"`
	Description string `bson:"problem_set_description"`
}

type ProblemSetProblems struct {
	Name string `bson:"problem_name"`
	Id   string `bson:"problem_id"`
}

type ProblemSet struct {
	Name        string                `bson:"name"`
	Description string                `bson:"description"`
	Problems    []*ProblemSetProblems `bson:"problems"`
}

func (p ProblemSet) TemplateProblems() [][]*ProblemSetProblems {
	var results [][]*ProblemSetProblems

	for row := 0; row <= len(p.Problems)/ProbColumns; row++ {
		var rowResult []*ProblemSetProblems
		for col := 0; col < ProbColumns && (row*ProbColumns+col) < len(p.Problems); col++ {
			rowResult = append(rowResult, p.Problems[row*HomeColumns+col])
		}
		results = append(results, rowResult)
	}
	return results
}

type Problem struct {
	Name        string    `bson:"problem_name"`
	Id          string    `bson:"problem_id"`
	Set         string    `bson:"problem_set"`
	NextProblem string    `bson:"next_problem"`
	PrevProblem string    `bson:"prev_problem"`
	Description string    `bson:"problem_description"`
	StartCode   string    `bson:"problem_start_code"`
	ProbTests   []*string `bson:"problem_tests"`
}

func (p Problem) DescriptionSplit() []string {
	return strings.Split(p.Description, "\n")
}

type DataType interface{}

type ProblemTest struct {
	Inputs []*DataType
	Output DataType
}

// Handlers
func homeHandler(w http.ResponseWriter, r *http.Request) {
	homeTemplates := append(baseTemplates, "templates/home.tmpl")
	t, err := template.ParseFiles(homeTemplates...)

	if err != nil {
		fmt.Println(err)
		return
	}

	// Get Problem Sets
	client, err := mongo.Connect(context.TODO(), "mongodb://localhost:27017")
	if err != nil {
		log.Println(err)
	}
	defer client.Disconnect(context.TODO())

	var problemSets []*ProblemSetOverview
	problemSetsColl := client.Database("goding-bat").Collection("problem_sets")
	cur, err := problemSetsColl.Find(context.TODO(), bson.D{})
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var probSet ProblemSetOverview
		err := cur.Decode(&probSet)
		if err != nil {
			log.Println(err)
		}
		problemSets = append(problemSets, &probSet)
	}

	if err := cur.Err(); err != nil {
		log.Println(err)
	}

	var results [][]*ProblemSetOverview
	for row := 0; row <= len(problemSets)/HomeColumns; row++ {
		var rowResult []*ProblemSetOverview
		for col := 0; col < HomeColumns && (row*HomeColumns+col) < len(problemSets); col++ {
			rowResult = append(rowResult, problemSets[row*HomeColumns+col])
		}
		results = append(results, rowResult)
	}

	err = t.ExecuteTemplate(w, "home.tmpl", results)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	// Get problem set
	vars := mux.Vars(r)
	probSetName := vars["setName"]

	// Check and Get Problem Set
	client, err := mongo.Connect(context.TODO(), "mongodb://localhost:27017")
	if err != nil {
		log.Println(err)
	}
	defer client.Disconnect(context.TODO())

	var problemSet ProblemSet
	problemMapColl := client.Database("goding-bat").Collection("problem_map")
	filter := bson.D{{"name", probSetName}}
	err = problemMapColl.FindOne(context.TODO(), filter).Decode(&problemSet)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	setTemplates := append(baseTemplates, "templates/problem_set.tmpl")
	t, err := template.ParseFiles(setTemplates...)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = t.ExecuteTemplate(w, "problem_set.tmpl", problemSet)
	if err != nil {
		fmt.Println(err.Error())
	}

}

func probHandler(w http.ResponseWriter, r *http.Request) {
	// Get problem info
	vars := mux.Vars(r)
	probId := vars["probId"]

	// Check and Get Problem Set
	client, err := mongo.Connect(context.TODO(), "mongodb://localhost:27017")
	if err != nil {
		log.Println(err)
	}
	defer client.Disconnect(context.TODO())

	var problem Problem
	problemColl := client.Database("goding-bat").Collection("problem_info")
	// TODO: change problem_id in mongo to bot include /prob/
	filter := bson.D{{"problem_id", "/prob/" + probId}}
	err = problemColl.FindOne(context.TODO(), filter).Decode(&problem)
	if err != nil {
		log.Println(err)
		http.NotFound(w, r)
		return
	}

	// Massage data
	//fmt.Println(problem)

	probTemplates := append(baseTemplates, "templates/problem.tmpl")
	t, err := template.ParseFiles(probTemplates...)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err != nil {
		fmt.Println(err.Error())
	}

	err = t.ExecuteTemplate(w, "problem.tmpl", problem)
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	routes := mux.NewRouter()
	routes.HandleFunc("/", homeHandler)
	//routes.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	routes.HandleFunc("/prob/{probId:p[0-9]+}", probHandler)
	routes.HandleFunc("/{setName:[a-zA-z]+\\-[0-9]+}", setHandler)
	http.Handle("/", routes)

	http.ListenAndServe(":8080", nil)
}
