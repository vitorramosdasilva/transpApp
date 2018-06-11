package main

import (
	"database/sql"
	"log"
	"net/http"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

//Tbclientes tabela clientes
type Tbclientes struct {
	ID        int
	Nome      string
	Email     string
	Telefone  string
	Descricao string
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "root"
	dbName := "bd_transpapp"

	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(127.0.0.1:3306)/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	log.Printf("conectado...")
	return db
}

var tmpl = template.Must(template.ParseGlob("form/*"))

//Index busca pelo index.......
func Index(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	selDB, err := db.Query("SELECT * FROM tb_clientes ORDER BY id DESC")
	if err != nil {
		panic(err.Error())
	}

	cli := Tbclientes{}
	var clientes []Tbclientes

	for selDB.Next() {

		var id int
		var nome, email string
		var telefone, descricao string

		err = selDB.Scan(&id, &nome, &email, &telefone, &descricao)

		if err != nil {
			panic(err.Error())
		}

		cli.ID = id
		cli.Nome = nome
		cli.Email = email
		cli.Telefone = telefone
		cli.Descricao = descricao

		clientes = append(clientes, cli)
	}

	log.Println(clientes)
	tmpl.ExecuteTemplate(w, "Index", clientes)
	defer db.Close()
}

//Show lista os clientes...
func Show(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	Nid := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM tb_clientes WHERE id=?", Nid)
	if err != nil {
		panic(err.Error())
	}
	cli := Tbclientes{}
	for selDB.Next() {
		var id int
		var nome, email string
		var telefone, descricao string

		err = selDB.Scan(&id, &nome, &email, &telefone, &descricao)
		if err != nil {
			panic(err.Error())
		}
		cli.ID = id
		cli.Nome = nome
		cli.Email = email
		cli.Telefone = telefone
		cli.Descricao = descricao
	}

	tmpl.ExecuteTemplate(w, "Show", cli)
	defer db.Close()
}

//New teplate para salvar....
func New(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "New", nil)
}

//Edit conteudo......
func Edit(w http.ResponseWriter, r *http.Request) {
	db := dbConn()

	Nid := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM tb_clientes WHERE id=?", Nid)
	if err != nil {
		panic(err.Error())
	}
	cli := Tbclientes{}
	for selDB.Next() {

		var id int
		var nome, email string
		var telefone, descricao string

		err = selDB.Scan(&id, &nome, &email, &telefone, &descricao)

		if err != nil {
			panic(err.Error())
		}
		cli.ID = id
		cli.Nome = nome
		cli.Email = email
		cli.Telefone = telefone
		cli.Descricao = descricao

	}

	tmpl.ExecuteTemplate(w, "Edit", cli)
	defer db.Close()
}

//Update atualizo cliente..
func Update(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		Nome := r.FormValue("Nome")
		Email := r.FormValue("Email")
		Telefone := r.FormValue("Telefone")
		Descricao := r.FormValue("Descricao")
		id := r.FormValue("uid")

		insForm, err := db.Prepare("UPDATE Tb_clientes SET nome=?, email=?,  telefone=?, descricao=? WHERE id=?")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(Nome, Email, Telefone, Descricao, id)
		log.Println("UPDATE: Name: " + Nome + " | " + Email + " | " + Telefone + " | " + Descricao + " | " + id)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

//Insert inseri novo conteudo....
func Insert(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {

		Nome := r.FormValue("Nome")
		Email := r.FormValue("Email")
		Telefone := r.FormValue("Telefone")
		Descricao := r.FormValue("Descricao")

		insForm, err := db.Prepare("INSERT INTO tb_clientes(nome, email, telefone, descricao) VALUES(?,?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(Nome, Email, Telefone, Descricao)
		log.Println("INSERT: campos: " + Nome + " | " + Email + " | " + Telefone + " | " + Descricao)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

//Delete deleto cliente.
func Delete(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	cli := r.URL.Query().Get("id")
	delForm, err := db.Prepare("DELETE FROM Tb_clientes WHERE id=?")
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(cli)
	log.Println("DELETE....")
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func main() {
	log.Println("Server started on: http://localhost:8080")
	http.HandleFunc("/", Index)
	http.HandleFunc("/show", Show)
	http.HandleFunc("/new", New)
	http.HandleFunc("/edit", Edit)
	http.HandleFunc("/insert", Insert)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/delete", Delete)
	http.ListenAndServe(":8080", nil)
}
