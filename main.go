package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const defaultDatabaseURL = "postgres://sourcegraph@localhost:5432/perms?sslmode=disable"

var databaseURL string

var db *sql.DB

func main() {
	databaseURL = os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}

	_db, err := newDB(databaseURL)
	if err != nil {
		log.Fatal("Getting db connector", err)
	}
	db = _db
	shutdown := initiateSetup(db)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(authCheckMiddleware)
	// r.Use(middleware.AllowContentType("application/json"))

	r.Get("/", rootHandler)

	r.Route("/batchchanges", func(r chi.Router) {
		r.Get("/", getBatchChangesHandler)
		r.Patch("/{batchChangeID}/share/{recipientUserID}/{action}", shareBatchChange)
		r.Post("/", createBatchChange)
		r.Get("/{batchChangeID}", getBatchChange)
	})

	go func() {
		<-c
		fmt.Println("shutting down ...")
		shutdown()
		os.Exit(0)
	}()

	log.Fatal(http.ListenAndServe(":3000", r))
}

func populateRolesWithPermissions(c *Controller, rs []*Role, ps []*Permission) error {
	fmt.Println("")
	fmt.Println(`We want the site administrator to have all permissions, however we want
the DEFAULT role to give full access to only notebooks,
view access to batch changes and view to code insights,
while the operator role will give full access to batch changes`)
	fmt.Println("")
	for _, r := range rs {
		// add all global permissions to the site admin role
		if r.Name == "SITE_ADMINISTRATOR" {
			fmt.Println("Adding all global permissions to the site admin role")
			for _, p := range ps {
				err := c.addPermissionToRole(r, p)
				if err != nil {
					return err
				}
			}
			continue
		}

		if r.Name == "DEFAULT" {
			fmt.Println("Adding permissions to the DEFAULT role ...")
			for _, p := range ps {
				if p.Namespace == "NOTEBOOKS" {
					err := c.addPermissionToRole(r, p)
					if err != nil {
						return err
					}
				}

				if p.Namespace == "CODEINSIGHTS" && p.Action == "READ" {
					err := c.addPermissionToRole(r, p)
					if err != nil {
						return err
					}
				}

				if p.Namespace == "BATCHCHANGES" && p.Action == "READ" {
					err := c.addPermissionToRole(r, p)
					if err != nil {
						return err
					}
				}
			}

			continue
		}

		if r.Name == "OPERATOR" {
			fmt.Println("Adding permissions to the OPERATOR role ...")
			for _, p := range ps {
				if p.Namespace == "BATCHCHANGES" {
					err := c.addPermissionToRole(r, p)
					if err != nil {
						return err
					}
				}
			}

			continue
		}
	}

	return nil
}

func initiateSetup(db *sql.DB) (teardown func()) {
	fmt.Println("============================= SETUP ========================================")
	teardown, err := setupDB(databaseURL)
	if err != nil {
		log.Fatalf("Migrations: %v", err)
	}

	c := &Controller{DB: db}

	rs, err := seedRoles(c)
	if err != nil {
		log.Fatalf("error while seeding users --> %s", err.Error())
	}
	fmt.Printf("Successfully seeded %d roles ... \n", len(rs))

	os, err := seedOrgs(c)
	if err != nil {
		log.Fatalf("error seeding orgs ---> %s", err.Error())
	}
	fmt.Printf("Successfully seeded %d orgs ....\n", len(os))

	s, err := ParseSchema()
	if err != nil {
		log.Fatal("error parsing schema")
	}

	ps, err := c.createPermissions(s)
	if err != nil {
		log.Fatalf("error creating permissions: %s", err.Error())
	}

	err = populateRolesWithPermissions(c, rs, ps)
	if err != nil {
		log.Fatalf("error populating roles with permissions: %s", err.Error())
	}

	users, err := seedUsers(c, rs...)
	if err != nil {
		log.Fatalf("error while seeding users --> %v", err)
	}
	fmt.Printf("Successfully seeded %d users ... \n", len(users))

	for _, bc := range mockBatchChanges {
		err = c.createBatchChange(&bc)
		if err != nil {
			log.Fatal("error creating batch change", err)
		}
	}

	for _, nb := range mockNotebooks {
		err = c.createNotebook(&nb)
		if err != nil {
			log.Fatal("error creating notebook", err)
		}
	}

	for _, ci := range mockCodeInsights {
		err = c.createCodeInsight(&ci)
		if err != nil {
			log.Fatal("error creating code insight", err)
		}
	}

	fmt.Println("============================= SETUP COMPLETE========================================")
	return
}
