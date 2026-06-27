package ports

// ProfileRepository define el contrato para el perfil profesional del actor.
// ADR-0021: separado de users para mantener el Core agnostico de dominio.
type ProfileRepository interface {
	Get(userID string) (Profile, error)
	Upsert(p Profile) error
}

// Profile representa el perfil profesional del medico.
type Profile struct {
	UserID            string
	TenantID          string
	Rubro             string
	NombreCompleto    string
	CedulaProfesional string
	Especialidad      string
}
