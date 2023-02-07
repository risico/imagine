package server

type ServerParams struct {
	hostname string
	port     int
}

func (p ServerParams) withDefaults() ServerParams {
	return p
}

type Server struct {
	params *ServerParams
}

func NewServer(params ServerParams) *Server {
	s := &Server{
		params: &params,
	}
	return s
}
