package main

import (
	"fmt"

	"github.com/el-bulk/backend/utils/logger"
	"github.com/jmoiron/sqlx"
)

// CustomerSeeded holds just enough for downstream seeders.
type CustomerSeeded struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

func seedCustomers(db *sqlx.DB) ([]CustomerSeeded, error) {
	logger.Info("👥 Seeding customers (30 — full field coverage)...")

	type Cust struct {
		FirstName   string
		LastName    string
		Email       string
		Phone       string
		IDNumber    string
		Address     string
		AvatarURL   string
		AuthProvider   string
		AuthProviderID string
	}

	customers := []Cust{
		// VIP customers (first 5 — get heavy activity)
		{"Sebastián", "Rodríguez", "sebastian.rodz@gmail.com", "3101234001", "1023456789",
			"Cra 15 # 80-22, Apto 601, Bogotá", "https://api.dicebear.com/7.x/avataaars/svg?seed=Sebastian",
			"google", "google-uid-0001"},
		{"Valentina", "Gómez", "vale.gomez@hotmail.com", "3201234002", "52345678",
			"Cl 72 # 4-12, Casa 3, Bogotá", "https://api.dicebear.com/7.x/avataaars/svg?seed=Valentina",
			"google", "google-uid-0002"},
		{"Andrés", "Martínez", "andres.mtcg@gmail.com", "3151234003", "80123456",
			"Av Suba # 111-40, Bogotá", "https://api.dicebear.com/7.x/avataaars/svg?seed=Andres",
			"google", "google-uid-0003"},
		{"Camila", "Torres", "camila.torres@outlook.com", "3001234004", "24567890",
			"Cl 100 # 15-30, Bogotá", "https://api.dicebear.com/7.x/avataaars/svg?seed=Camila",
			"google", "google-uid-0004"},
		{"Felipe", "Herrera", "fherrera.tcg@gmail.com", "3211234005", "71234567",
			"Cra 30 # 65-10, Medellín", "https://api.dicebear.com/7.x/avataaars/svg?seed=Felipe",
			"google", "google-uid-0005"},

		// Regular customers with Google OAuth (next 10)
		{"Laura", "Castillo", "laura.castillo@gmail.com", "3051234006", "1042345678",
			"Cl 45 # 22-18, Bogotá", "https://api.dicebear.com/7.x/avataaars/svg?seed=Laura",
			"google", "google-uid-0006"},
		{"Juan Carlos", "Moreno", "juanc.moreno@gmail.com", "3171234007", "79012345",
			"Cra 7 # 32-45, Bogotá", "", "google", "google-uid-0007"},
		{"Diana", "Peña", "diana.pena.tcg@gmail.com", "3061234008", "36789012",
			"Cl 80 # 11-22, Bogotá", "", "google", "google-uid-0008"},
		{"Nicolás", "García", "nicolas.garcia@icloud.com", "3121234009", "1014567890",
			"Cra 50 # 90-15, Bogotá", "", "google", "google-uid-0009"},
		{"María José", "López", "mariajose.lopez@gmail.com", "3111234010", "46789012",
			"Av. El Dorado # 69-76, Bogotá", "", "google", "google-uid-0010"},
		{"David", "Ramírez", "david.ramirez@gmail.com", "3141234011", "80901234",
			"Cl 26 # 59-41, Bogotá", "", "google", "google-uid-0011"},
		{"Catalina", "Vargas", "catalina.vargas@gmail.com", "3091234012", "52890123",
			"Cra 19 # 84-14, Bogotá", "", "google", "google-uid-0012"},
		{"Esteban", "Jiménez", "esteban.j@gmail.com", "3131234013", "1020123456",
			"Cl 53 # 28-55, Bogotá", "", "google", "google-uid-0013"},
		{"Juliana", "Sánchez", "juli.sanchez@gmail.com", "3161234014", "23456789",
			"Cra 11 # 93-20, Bogotá", "", "google", "google-uid-0014"},
		{"Santiago", "Díaz", "santiago.diaz.magic@gmail.com", "3181234015", "71890123",
			"Autopista Norte # 155-50, Bogotá", "", "google", "google-uid-0015"},

		// Guest customers (no OAuth — from checkout form)
		{"Carlos", "Ospina", "carlos.ospina@yahoo.com", "3221234016", "12345678", "Cra 3 # 10-20, Cali", "", "", ""},
		{"Marcela", "Ruiz", "m.ruiz.gaming@gmail.com", "3231234017", "29012345", "Cl 5 # 15-30, Cali", "", "", ""},
		{"Alejandro", "Muñoz", "alejandro.pokemon@gmail.com", "3241234018", "80234567", "Av 4N # 23-45, Cali", "", "", ""},
		{"Paula", "Rojas", "paula.rojas@hotmail.com", "3251234019", "39012345", "Cra 1 # 5-10, Medellín", "", "", ""},
		{"Miguel", "Suárez", "miguel.yugioh@outlook.com", "3261234020", "71345678", "Cl Palacé # 40-50, Medellín", "", "", ""},
		{"Natalia", "Correa", "natalia.correa@gmail.com", "3271234021", "52012345", "Cra 43A # 3-63, Medellín", "", "", ""},
		{"Mateo", "Restrepo", "mateo.rest@gmail.com", "3281234022", "1019012345", "Cl 10 # 4-50, Medellín", "", "", ""},
		{"Isabella", "Mejía", "isa.mejia.tcg@gmail.com", "3291234023", "43567890", "Cl 72 # 64-60, Barranquilla", "", "", ""},
		{"Tomás", "Ríos", "tomas.rios@gmail.com", "3101234024", "72678901", "Cra 46 # 67-30, Barranquilla", "", "", ""},
		{"Daniela", "Guzmán", "dani.guzman@gmail.com", "3201234025", "32890123", "Cra 54 # 75-70, Barranquilla", "", "", ""},
		{"Roberto", "Acosta", "r.acosta.magic@gmail.com", "3151234026", "79567890", "Cl 30 # 20-10, Bucaramanga", "", "", ""},
		{"Ana María", "Pinto", "anapinto.lorcana@gmail.com", "3001234027", "46234567", "Cra 27 # 52-20, Bucaramanga", "", "", ""},
		{"Gabriel", "Mendoza", "gabriel.m.tcg@gmail.com", "3211234028", "1030234567", "Cl 16 # 6-66, Cúcuta", "", "", ""},
		{"Valeria", "Cárdenas", "vale.cardenas@gmail.com", "3051234029", "39456789", "Av 0 # 11-26, Cúcuta", "", "", ""},
		{"Emilio", "Franco", "emilio.franco.mtg@gmail.com", "3171234030", "71456789", "Cl 15 # 10-30, Cartagena", "", "", ""},
	}

	var seeded []CustomerSeeded
	for i, c := range customers {
		createdAt := daysAgo(randInt(5, 120))

		var idNum, addr, avatar *string
		if c.IDNumber != "" {
			idNum = &c.IDNumber
		}
		if c.Address != "" {
			addr = &c.Address
		}
		if c.AvatarURL != "" {
			avatar = &c.AvatarURL
		}

		var authProvider, authProviderID *string
		if c.AuthProvider != "" {
			authProvider = &c.AuthProvider
			authProviderID = &c.AuthProviderID
		}

		var id string
		err := db.QueryRow(`
			INSERT INTO customer (
				first_name, last_name, email, phone,
				id_number, address, avatar_url,
				auth_provider, auth_provider_id, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT (phone) DO UPDATE SET
				first_name = EXCLUDED.first_name,
				last_name = EXCLUDED.last_name,
				email = EXCLUDED.email
			RETURNING id
		`,
			c.FirstName, c.LastName, c.Email, c.Phone,
			idNum, addr, avatar,
			authProvider, authProviderID, createdAt,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to seed customer '%s %s': %w", c.FirstName, c.LastName, err)
		}

		// Seed customer_auth separately for first 15 (OAuth users)
		if c.AuthProvider != "" {
			db.Exec(`
				INSERT INTO customer_auth (customer_id, provider, provider_id)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, id, c.AuthProvider, c.AuthProviderID)

			// First 5 VIPs also get Facebook
			if i < 5 {
				db.Exec(`
					INSERT INTO customer_auth (customer_id, provider, provider_id)
					VALUES ($1, 'facebook', $2)
					ON CONFLICT DO NOTHING
				`, id, fmt.Sprintf("fb-uid-%04d", i+1))
			}
		}

		seeded = append(seeded, CustomerSeeded{
			ID:        id,
			FirstName: c.FirstName,
			LastName:  c.LastName,
			Email:     c.Email,
			Phone:     c.Phone,
		})
	}

	logger.Info("✅ %d customers seeded", len(seeded))
	return seeded, nil
}
