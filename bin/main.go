package main

import (
	"database/sql"
	"fmt"
	//"text/template"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	//"time"

	"github.com/gorilla/sessions"
	//"github.com/vjeantet/jodaTime"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var encryptionKey = "something-very-secret"
var loggedUserSession = sessions.NewCookieStore([]byte(encryptionKey))
var p Tbparceiros
var tmpl = template.Must(template.ParseGlob("pages/*"))

//var conditionsMap = map[string]interface{}{}

func init() {

	loggedUserSession.Options = &sessions.Options{
		// change domain to match your machine. Can be localhost
		// IF the Domain name doesn't match, your session will be EMPTY!
		Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 0.4, // 1 hours
		HttpOnly: true,
	}
}

//Tbclientes tabela clientes
type Tbclientes struct {
	ID        int64
	Nome      string
	Email     string
	Telefone  string
	Descricao string
	Mensagem  string
	//DataCadastro time.Time
	DataCadastro string
}

//Tbparceiros tabela Tbtrabalheconosco
type Tbparceiros struct {
	ID          int64
	Nome        string
	Email       string
	Password    string
	Telefone1   string
	Telefone2   string
	TipoVeiculo string
	Mensagem    string
	//DataCadastro time.Time
	DataCadastro string
	TipoAcesso   string
	Logado       bool
}

//Tbenderecosparceiros tabela Tbtrabalheendereco
type Tbenderecosparceiros struct {
	ID          int64
	Cep         string
	Rua         string
	Numero      string
	Bairro      string
	Complemento string
	Cidade      string
	Uf          string
	idparc      Tbparceiros
	Mensagem    string
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "root"
	dbName := "bd_transpapp"
	dbCharset := "utf8"

	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(127.0.0.1:3306)/"+dbName+"?charset="+dbCharset+"&parseTime=true")
	//db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(127.0.0.1:3306)/"+dbName+"?sql_mode=TRADITIONAL")
	if err != nil {
		log.Fatal("Cannot open DB connection", err)
		panic(err.Error())
	}
	log.Printf("conectado...")
	return db
}

//InsertContato inseri contato....
func InsertContato(w http.ResponseWriter, r *http.Request) {
	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")

	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username : ", session.Values["username"])
	conditionsMap["Username"] = session.Values["username"]

	db := dbConn()
	if r.Method == "POST" {

		Nome := r.FormValue("Nome")
		Email := r.FormValue("Email")
		Telefone := r.FormValue("Telefone")
		Descricao := r.FormValue("Descricao")

		insForm, err := db.Prepare("INSERT INTO tb_clientes(nome, email, telefone, descricao) VALUES(?,?,?,?)")
		if err != nil {
			log.Fatal("InsertContato erro: ", err)
			panic(err.Error())
		}
		insForm.Exec(Nome, Email, Telefone, Descricao)
		log.Println("INSERT: campos: " + Nome + " | " + Email + " | " + Telefone + " | " + Descricao)

		conditionsMap["Mensagem"] = "Cadastro Efetuado com  sucesso, retornaremos em breve."
		tmpl.ExecuteTemplate(w, "contact.html", conditionsMap)
		defer db.Close()
	} else {
		http.Redirect(w, r, "/404", 301)
	}
	//return
}

//InsertTrabalheConosco inseri TrabalheConosco....
func InsertTrabalheConosco(w http.ResponseWriter, r *http.Request) {

	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")

	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}

	existe := ExisteUsuario(w, r)
	if r.Method == "POST" {
		if !existe {

			db := dbConn()
			Nome := r.FormValue("Nome")
			Email := r.FormValue("Email")
			Password := r.FormValue("Password")
			hash, _ := HashPassword(Password)
			Telefone1 := r.FormValue("Telefone1")
			Telefone2 := r.FormValue("Telefone2")
			TipoVeiculo := r.FormValue("TipoVeiculo")
			TipoAcesso := "usuario"

			insForm, err := db.Prepare("INSERT INTO Tb_parceiros(nome, email, password, telefone1, " +
				"telefone2, tipoveiculo, tipoAcesso) VALUES(?,?,?,?,?,?,?)")

			if err != nil {
				log.Fatal("Sintax InsertTrabalheConosco Error: ", err)
				panic(err.Error())
			}

			res, err := insForm.Exec(Nome, Email, hash, Telefone1, Telefone2, TipoVeiculo, TipoAcesso)
			log.Println("INSERT: campos: " + Nome + " | " + Email + " | " + hash + " | " + Telefone1 + " | " + Telefone2 + " | " + TipoVeiculo + " | " + TipoAcesso)

			if err != nil {
				log.Fatal("InsertTrabalheConosco Error: ", err)
			}

			conditionsMap["ID"], err = res.LastInsertId()
			conditionsMap["Mensagem"] = "Cadastro Dados Pessoais Efetuado com  sucesso, Faltam os Dados do Endereço"
			log.Println("Session name : ", session.Name())
			log.Println("Username : ", session.Values["username"])
			conditionsMap["Username"] = session.Values["username"]
			session.Values["tipoAcesso"] = "usuario"
			conditionsMap["Username"] = session.Values["tipoAcesso"]
			tmpl.ExecuteTemplate(w, "trabalheEndereco.html", conditionsMap)
			fmt.Printf("Inserted row: %d", conditionsMap["ID"])
			defer db.Close()

		} else {
			conditionsMap["Mensagem"] = "Email já existe em nosso banco de dados, por favor faça login."
			tmpl.ExecuteTemplate(w, "login.html", conditionsMap)
		}
	} else {
		http.Redirect(w, r, "/404", 301)
	}

}

//InsertTrabalheEndereco inseri endereco....
func InsertTrabalheEndereco(w http.ResponseWriter, r *http.Request) {

	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username : ", session.Values["username"])
	conditionsMap["Username"] = session.Values["username"]

	db := dbConn()

	if r.Method == "POST" {

		Cep := r.FormValue("Cep")
		Rua := r.FormValue("Rua")
		Numero := r.FormValue("Numero")
		Bairro := r.FormValue("Bairro")
		Complemento := r.FormValue("Complemento")
		Cidade := r.FormValue("Cidade")
		Uf := r.FormValue("Uf")
		ID := string(r.FormValue("Id_parc"))

		if ID == "0" || ID == "" {
			conditionsMap["Mensagem"] = "Efetue Cadastro Primeiramente dos Dados Pessoais"
			tmpl.ExecuteTemplate(w, "trabalheConosco.html", conditionsMap)
			return
		}

		insForm, err := db.Prepare("INSERT INTO tb_enderecos_parceiros(cep, rua, numero, bairro, " +
			"complemento, Cidade, uf,Id_parc) VALUES(?,?,?,?,?,?,?,?)")
		log.Println("INSERT: campos: " + Cep + " | " + Rua + " | " + Numero + " | " + Bairro + " | " + Complemento + " | " + Cidade + " | " + Uf + " | " + ID)

		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(Cep, Rua, Numero, Bairro, Complemento, Cidade, Uf, ID)
		log.Println("INSERT: campos: " + Cep + " | " + Rua + " | " + Numero + " | " + Bairro + " | " + Complemento + " | " + Cidade + " | " + Uf + " | " + ID)

		conditionsMap["Mensagem"] = "Cadastro de Endereço Efetuado com  sucesso"
		conditionsMap["ID"], err = strconv.ParseInt(ID, 10, 64)

		tmpl.ExecuteTemplate(w, "trabalheEndereco.html", conditionsMap)
		fmt.Printf("Inserted row: %d", conditionsMap["ID"])
		defer db.Close()

	} else {
		http.Redirect(w, r, "/404", 301)
	}

}

//UpdateTrabalheEndereco atualizo cliente..
func UpdateTrabalheEndereco(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		Cep := r.FormValue("Cep")
		Rua := r.FormValue("Rua")
		Numero := r.FormValue("Numero")
		Bairro := r.FormValue("Bairro")
		Complemento := r.FormValue("Complemento")
		Cidade := r.FormValue("Cidade")
		Uf := r.FormValue("Uf")
		ID := string(r.FormValue("Id_parc"))
		log.Println("INSERT: campos: " + Cep + " | " + Rua + " | " + Numero + " | " + Bairro + " | " + Complemento + " | " + Cidade + " | " + Uf + "|" + ID)

		insForm, err := db.Prepare("UPDATE tb_enderecos_parceiros SET cep=?, rua=?,  numero=?, bairro=?," +
			"complemento=?, cidade=?,  uf=?,  id_parc=? WHERE id=?")

		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(Cep, Rua, Numero, Bairro, Complemento, Cidade, Uf, ID)
		log.Println("INSERT: campos: " + Cep + " | " + Rua + " | " + Numero + " | " + Bairro + " | " + Complemento + " | " + Cidade + " | " + Uf)

	} else {
		http.Redirect(w, r, "/404", 301)
		return
	}

	defer db.Close()
	tmpl.ExecuteTemplate(w, "login.html", p)

}

//Logout desloga usuarios...
func Logout(w http.ResponseWriter, r *http.Request) {
	//read from session
	session, _ := loggedUserSession.Get(r, "authenticated-user-session")
	// remove the username
	session.Values["username"] = nil
	session.Values["tipoAcesso"] = nil
	err := session.Save(r, w)
	if err != nil {
		log.Println(err)
	}
	log.Println("Logged out!")
	tmpl.ExecuteTemplate(w, "login.html", nil)
}

//Logar loga em sistema.......
func Logar(w http.ResponseWriter, r *http.Request) {
	conditionsMap := map[string]interface{}{}
	// check if session is active
	session, _ := loggedUserSession.Get(r, "authenticated-user-session")
	db := dbConn()

	if session != nil {
		conditionsMap["Username"] = session.Values["username"]
	}

	if r.Method == "POST" {

		email := r.FormValue("Email")
		senha := r.FormValue("Password")
		err := db.QueryRow("select ID, Email, Password, Nome, TipoAcesso from Tb_parceiros WHERE email=?", email).Scan(&p.ID, &p.Email, &p.Password, &p.Nome, &p.TipoAcesso)
		validaPass := CheckPasswordHash(senha, p.Password)

		if err != nil || !validaPass {
			log.Println("Either username or password is wrong")
			conditionsMap["LoginError"] = true
			conditionsMap["Mensagem"] = "Email ou Senha Incorretos"
			tmpl.ExecuteTemplate(w, "login.html", conditionsMap)
			defer db.Close()
			return
		}

		username := p.Nome
		log.Println("Logged in :", username)
		conditionsMap["Username"] = username
		conditionsMap["LoginError"] = false

		// create a new session and redirect to dashboard
		session, _ := loggedUserSession.New(r, "authenticated-user-session")
		session.Values["username"] = username
		session.Values["tipoAcesso"] = p.TipoAcesso
		conditionsMap["tipoAcesso"] = session.Values["tipoAcesso"]
		err = session.Save(r, w)

		if err != nil {
			log.Println(err)
		}

		defer db.Close()
		//consulta aqui........

		if p.TipoAcesso == "adm" {

			db := dbConn()
			selDB, err := db.Query("SELECT * FROM tb_clientes ORDER BY id DESC")
			if err != nil {
				panic(err.Error())
			}

			cli := Tbclientes{}
			var clientes []Tbclientes

			for selDB.Next() {

				var id int64
				var nome, email, telefone, descricao string
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
			tmpl.ExecuteTemplate(w, "painel.html", clientes)

		} else {
			tmpl.ExecuteTemplate(w, "painelCliente.html", conditionsMap)
		}

		defer db.Close()
		return

	}
	http.Redirect(w, r, "/404", 301)

}

//ExisteUsuario atualizo ListaCidades......
func ExisteUsuario(w http.ResponseWriter, r *http.Request) bool {

	db := dbConn()

	Nid := r.FormValue("Email")
	var nome string
	err := db.QueryRow("SELECT email FROM Tb_parceiros WHERE email=?", Nid).Scan(&nome)
	//selDB, err := db.Query("SELECT email FROM Tb_parceiros WHERE email=?", Nid)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("No user with that ID.")
		return false
	case err != nil:
		log.Fatal(err)
		return false
	default:
		fmt.Printf("Username is %s\n", nome)
	}

	defer db.Close()
	return true
	//tmpl.ExecuteTemplate(w, "Show", cli)
}

//About é a About......
func About(w http.ResponseWriter, r *http.Request) {
	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")

	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username About : ", session.Values["username"])

	conditionsMap["Username"] = session.Values["username"]

	if session.Values["tipoAcesso"] == "adm" {
		conditionsMap["tipoAcesso"] = session.Values["tipoAcesso"]
	}

	if conditionsMap["Username"] != "" && conditionsMap["Username"] != nil {
		tmpl.ExecuteTemplate(w, "about.html", conditionsMap)
		log.Println("Contém username : ", session.Values["username"])
	} else {
		tmpl.ExecuteTemplate(w, "about.html", nil)
		log.Println("Sem username : ", session.Values["username"])
	}

}

//Home esta é a home......
func Home(w http.ResponseWriter, r *http.Request) {
	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username Home : ", session.Values["username"])
	conditionsMap["Username"] = session.Values["username"]

	if session.Values["tipoAcesso"] == "adm" {
		conditionsMap["tipoAcesso"] = session.Values["tipoAcesso"]
	}

	if conditionsMap["Username"] != "" && conditionsMap["Username"] != nil {
		tmpl.ExecuteTemplate(w, "home.html", conditionsMap)
		log.Println("Contém Username Home : ", session.Values["username"])
	} else {
		tmpl.ExecuteTemplate(w, "home.html", nil)
		log.Println("Sem Username Home : ", session.Values["username"])
	}

}

//Contact esta é a home......
func Contact(w http.ResponseWriter, r *http.Request) {
	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username Contact : ", session.Values["username"])
	conditionsMap["Username"] = session.Values["username"]

	if session.Values["tipoAcesso"] == "adm" {
		conditionsMap["tipoAcesso"] = session.Values["tipoAcesso"]
	}

	if conditionsMap["Username"] != "" && conditionsMap["Username"] != nil {
		tmpl.ExecuteTemplate(w, "contact.html", conditionsMap)
		log.Println("Contém Username Contact : ", session.Values["username"])
	} else {
		tmpl.ExecuteTemplate(w, "contact.html", nil)
		log.Println("Sem Username Contact : ", session.Values["username"])
	}

}

/*
func EstaLogado(session string) bool {
	conditionsMap := map[string]interface{}{}
	if session.Values["username"] != nil {
		return true
	}
	return false
}

//TipoAcesso verifica tipo acesso.....
func TipoAcessoAdm(session string) bool {
	if session.Values["tipoAcesso"] == "adm" {
		return true
	}
	return false
}*/

//ContactInsert esta é a home......
func ContactInsert(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "contact.html", nil)
}

//Login logar no sistema......
func Login(w http.ResponseWriter, r *http.Request) {
	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")
	if err != nil {
		log.Println("Não há dados na sessão login", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username Login : ", session.Values["username"])
	conditionsMap["Username"] = session.Values["username"]

	if session.Values["tipoAcesso"] == "adm" {
		conditionsMap["tipoAcesso"] = session.Values["tipoAcesso"]
	}

	if conditionsMap["Username"] != "" && conditionsMap["Username"] != nil {
		tmpl.ExecuteTemplate(w, "login.html", conditionsMap)
	} else {
		tmpl.ExecuteTemplate(w, "login.html", nil)
	}
}

//LoginExiste cadastro no sistema......
func LoginExiste(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "loginExiste.html", nil)
}

//TrabalheConosco cadastro transportadores...
func TrabalheConosco(w http.ResponseWriter, r *http.Request) {
	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username trabalheConosco : ", session.Values["username"])
	conditionsMap["Username"] = session.Values["username"]

	if session.Values["tipoAcesso"] == "adm" {
		conditionsMap["tipoAcesso"] = session.Values["tipoAcesso"]
	}

	if conditionsMap["Username"] != "" && conditionsMap["Username"] != nil {
		tmpl.ExecuteTemplate(w, "trabalheConosco.html", conditionsMap)
	} else {
		tmpl.ExecuteTemplate(w, "trabalheConosco.html", nil)
	}
}

//TrabalheEndereco cadastro transportadores...
func TrabalheEndereco(w http.ResponseWriter, r *http.Request) {
	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")
	if err != nil {
		log.Println("Sem dados na sessão endereco!", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username TrabalheEndereco : ", session.Values["username"])
	conditionsMap["Username"] = session.Values["username"]

	if session.Values["tipoAcesso"] == "adm" {
		conditionsMap["tipoAcesso"] = session.Values["tipoAcesso"]
	}

	if conditionsMap["Username"] != "" && conditionsMap["Username"] != nil {
		tmpl.ExecuteTemplate(w, "trabalheEndereco.html", conditionsMap)
	} else {
		tmpl.ExecuteTemplate(w, "trabalheEndereco.html", nil)
	}
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

		var id int64
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

	tmpl.ExecuteTemplate(w, "edit", cli)
	defer db.Close()
}

//EdParceiro conteudo......
func EdParceiro(w http.ResponseWriter, r *http.Request) {
	db := dbConn()

	Nid := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT id, nome, email, telefone1, telefone2, tipoveiculo, tipoAcesso, DataCadastro FROM tb_parceiros WHERE id=?", Nid)
	if err != nil {
		panic(err.Error())
	}
	parca := Tbparceiros{}

	for selDB.Next() {

		var id int64
		var nome, email, telefone1, telefone2, tipoveiculo, tipoAcesso, DataCadastro string
		err = selDB.Scan(&id, &nome, &email, &telefone1, &telefone2, &tipoveiculo, &tipoAcesso, &DataCadastro)

		if err != nil {
			panic(err.Error())
		}

		parca.ID = id
		parca.Nome = nome
		parca.Email = email
		parca.Telefone1 = telefone1
		parca.Telefone2 = telefone2
		parca.TipoVeiculo = tipoveiculo
		parca.TipoAcesso = tipoAcesso
		parca.DataCadastro = DataCadastro		

	}

	tmpl.ExecuteTemplate(w, "edParceiro", parca)
	defer db.Close()
}

//ShParceiro lista os clientes...
func ShParceiro(w http.ResponseWriter, r *http.Request) {

	db := dbConn()
	Nid := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT id, nome, email, telefone1, telefone2, tipoveiculo, tipoAcesso, DataCadastro FROM tb_parceiros WHERE id=?", Nid)
	if err != nil {
		panic(err.Error())
	}

	var parca Tbparceiros

	for selDB.Next() {

		var id int64
		var nome, email, telefone1, telefone2, tipoveiculo, tipoAcesso, DataCadastro string
		

		err = selDB.Scan(&id, &nome, &email, &telefone1, &telefone2, &tipoveiculo, &tipoAcesso, &DataCadastro)

		if err != nil {
			panic(err.Error())
		}

		parca.ID = id
		parca.Nome = nome
		parca.Email = email
		parca.Telefone1 = telefone1
		parca.Telefone2 = telefone2
		parca.TipoVeiculo = tipoveiculo
		parca.TipoAcesso = tipoAcesso		
		parca.DataCadastro = DataCadastro
		
	}

	tmpl.ExecuteTemplate(w, "shParceiro", parca)
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

	var cli Tbclientes

	for selDB.Next() {

		var id int64
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

	tmpl.ExecuteTemplate(w, "show", cli)
	defer db.Close()
}

//DelParceiro deleto cliente.
func DelParceiro(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	parca := r.URL.Query().Get("id")
	delForm, err := db.Prepare("DELETE FROM Tb_parceiros WHERE id=?")
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(parca)
	log.Println("DELETE PARCEIRO....")
	defer db.Close()
	http.Redirect(w, r, "/listPainelParceiro", 301)
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
	http.Redirect(w, r, "/listaPainel", 301)
}

//ListaPainel busca pelo index.......
func ListaPainel(w http.ResponseWriter, r *http.Request) {

	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username : ", session.Values["username"])
	conditionsMap["Username"] = session.Values["username"]
	conditionsMap["tipoAcesso"] = session.Values["tipoAcesso"]

	if conditionsMap["tipoAcesso"] == "adm" && conditionsMap["Username"] != nil && conditionsMap["Username"] != "" {
		db := dbConn()
		selDB, err := db.Query("SELECT * FROM tb_clientes ORDER BY id DESC")
		if err != nil {
			panic(err.Error())
		}

		var clientes []Tbclientes
		cli := Tbclientes{}

		for selDB.Next() {

			var id int64
			var nome, email, telefone, descricao string
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
		tmpl.ExecuteTemplate(w, "painel.html", clientes)
		defer db.Close()
		return
	}
	http.Redirect(w, r, "/404", 301)

}

//ListPainelParceiro lisra todos parceiros....
func ListPainelParceiro(w http.ResponseWriter, r *http.Request) {

	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username : ", session.Values["username"])
	conditionsMap["Username"] = session.Values["username"]
	conditionsMap["tipoAcesso"] = session.Values["tipoAcesso"]

	if conditionsMap["tipoAcesso"] == "adm" && conditionsMap["Username"] != nil && conditionsMap["Username"] != "" {
		db := dbConn()
		selDB, err := db.Query("SELECT id, nome, email, telefone1, telefone2, tipoveiculo, tipoAcesso, DataCadastro FROM Tb_parceiros ORDER BY id DESC")

		if err != nil {
			panic(err.Error())
		}

		var parceiros []Tbparceiros
		parca := Tbparceiros{}

		for selDB.Next() {
			var id int64
			var nome, email, telefone1, telefone2, tipoveiculo, tipoAcesso, DataCadastro string
			err = selDB.Scan(&id, &nome, &email, &telefone1, &telefone2, &tipoveiculo, &tipoAcesso, &DataCadastro)

			if err != nil {
				panic(err.Error())
			}

			parca.ID = id
			parca.Nome = nome
			parca.Email = email
			parca.Telefone1 = telefone1
			parca.Telefone2 = telefone2
			parca.TipoVeiculo = tipoveiculo
			parca.TipoAcesso = tipoAcesso
			parca.DataCadastro = DataCadastro
			parceiros = append(parceiros, parca)
		}

		log.Println(parceiros)
		tmpl.ExecuteTemplate(w, "painelParceiro.html", parceiros)
		defer db.Close()
		return
	}
	http.Redirect(w, r, "/404", 301)

}

//Painel atualizo ListaCidades....
func Painel(w http.ResponseWriter, r *http.Request) {

	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username : ", session.Values["username"])
	conditionsMap["Username"] = session.Values["username"]

	if session.Values["tipoAcesso"] == "adm" && conditionsMap["Username"] != nil && conditionsMap["Username"] != "" {
		log.Println("Contém Username : ", session.Values["username"])
		http.Redirect(w, r, "/listaPainel", 301)
		return

	}

	tmpl.ExecuteTemplate(w, "404.html", nil)
	log.Println("Sem Username : ", session.Values["username"])

}

//PainelParceiro atualizo ListaCidades....
func PainelParceiro(w http.ResponseWriter, r *http.Request) {

	conditionsMap := map[string]interface{}{}
	session, err := loggedUserSession.Get(r, "authenticated-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}

	log.Println("Session name : ", session.Name())
	log.Println("Username : ", session.Values["username"])
	conditionsMap["Username"] = session.Values["username"]

	if session.Values["tipoAcesso"] == "adm" && conditionsMap["Username"] != nil && conditionsMap["Username"] != "" {
		log.Println("Contém Username : ", session.Values["username"])
		http.Redirect(w, r, "/listPainelParceiro", 301)
		return

	}

	tmpl.ExecuteTemplate(w, "404.html", nil)
	log.Println("Sem Username : ", session.Values["username"])

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
		log.Println("UPDATE....")
		defer db.Close()
		http.Redirect(w, r, "/listaPainel", 301)

	} else {
		http.Redirect(w, r, "/404", 301)
		//p.Mensagem = "Dados Editados com sucesso"
		//tmpl.ExecuteTemplate(w, "painel.html", p)

	}

}

//UpParceiro atualizo cliente..
func UpParceiro(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {

		db := dbConn()
		id := r.FormValue("uid")
		Nome := r.FormValue("Nome")
		Email := r.FormValue("Email")
		Telefone1 := r.FormValue("Telefone1")
		Telefone2 := r.FormValue("Telefone2")
		TipoVeiculo := r.FormValue("TipoVeiculo")
		TipoAcesso := r.FormValue("TipoAcesso")

		insForm, err := db.Prepare("UPDATE Tb_parceiros SET nome=?, email=?,  telefone1=?, " +
			"telefone2=?, TipoVeiculo=?, TipoAcesso=? WHERE id=?")

		if err != nil {
			panic(err.Error())
		}

		insForm.Exec(Nome, Email, Telefone1, Telefone2, TipoVeiculo, TipoAcesso, id)
		log.Println("Update: Parceiro: " + Nome + " | " + Email + " | " + Telefone1 + " | " + Telefone2 + " | " + TipoVeiculo + " | " + TipoAcesso + " | " + id)
		defer db.Close()
		http.Redirect(w, r, "/listPainelParceiro", 301)

	} else {
		http.Redirect(w, r, "/404", 301)

	}

}

//LembrarSenha esta é a LembrarSenha......
func LembrarSenha(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "lembrarSenha.html", nil)
}

//Erro404 esta é a Erro404......
func Erro404(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "404.html", nil)
}

//HashPassword transforma a senha em bytes................
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

//CheckPasswordHash verififica se a senha esta em bytes..........
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

//DualAxes pargina teste......
func DualAxes(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "dual-axes.html", nil)
}

func main() {

	var rootdir string
	rootdir, err := os.Getwd()

	if err != nil {
		rootdir = "No dice"
	}

	// Handler for anything pointing to images, css e js....
	http.Handle("/Public/bs4/img/", http.StripPrefix("/Public/bs4/img",
		http.FileServer(http.Dir(path.Join(rootdir, "Public/bs4/img/")))))
	http.Handle("/Public/bs4/css/", http.StripPrefix("/Public/bs4/css",
		http.FileServer(http.Dir(path.Join(rootdir, "Public/bs4/css/")))))
	http.Handle("/Public/bs4/js/", http.StripPrefix("/Public/bs4/js",
		http.FileServer(http.Dir(path.Join(rootdir, "Public/bs4/js/")))))

	log.Println("Server started on: http://localhost:8080")
	http.HandleFunc("/", Erro404)
	http.HandleFunc("/home", Home)
	http.HandleFunc("/about", About)
	http.HandleFunc("/contact", Contact)
	http.HandleFunc("/show", Show)
	http.HandleFunc("/shParceiro", ShParceiro)
	http.HandleFunc("/contactInsert", ContactInsert)
	http.HandleFunc("/loginExiste", LoginExiste)
	http.HandleFunc("/logout", Logout)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/logar", Logar)
	http.HandleFunc("/painel", Painel)
	http.HandleFunc("/painelParceiro", PainelParceiro)
	http.HandleFunc("/listaPainel", ListaPainel)
	http.HandleFunc("/listPainelParceiro", ListPainelParceiro)
	http.HandleFunc("/trabalheConosco", TrabalheConosco)
	http.HandleFunc("/trabalheEndereco", TrabalheEndereco)
	http.HandleFunc("/lembrarSenha", LembrarSenha)
	http.HandleFunc("/insertContato", InsertContato)
	http.HandleFunc("/insertTrabalheConosco", InsertTrabalheConosco)
	http.HandleFunc("/insertTrabalheEndereco", InsertTrabalheEndereco)
	http.HandleFunc("/updateTrabalheEndereco", UpdateTrabalheEndereco)
	http.HandleFunc("/edit", Edit)
	http.HandleFunc("/edParceiro", EdParceiro)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/upParceiro", UpParceiro)
	http.HandleFunc("/delete", Delete)
	http.HandleFunc("/delParceiro", DelParceiro)
	http.HandleFunc("/dualAxes", DualAxes)
	http.ListenAndServe(":8080", nil)
}
