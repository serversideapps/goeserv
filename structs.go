package main

type User struct {
	Handle   string   `json:"handle"`
	Rating   float64  `json:"rating"`
	Rd       float64  `json:"rd"`
	Email    string   `json:"email"`
}

type Presentation struct {
	Id                 string     `json:"id"`
	Title              string     `json:"title"`
	Owner              string     `json:"owner"`
	Pgn                string     `json:"pgn"`
	Currentlinealgeb   string     `json:"currentlinealgeb"`
	Version            int        `json:"version"`
	Book               Book       `json:"book"`
	Flip               bool       `json:"flip"`
	Candelete          string     `json:"candelete"`
	Canedit            string     `json:"canedit"`
}

func (pr *Presentation) checksanity() {
	pr.Book.checksanity()
}

type Game struct {
	Presentationid     string  `json:"presentationid"`
	Presentationtitle  string  `json:"presentationtitle"`
}

type GameWithPresentation struct {
	Presentationid     string        `json:"presentationid"`
	Presentationtitle  string        `json:"presentationtitle"`
	Presentation       Presentation  `json:"presentation"`
}

type StorePresentationMessage struct {
	Presid  string                `json:"presid"`
	Presg   GameWithPresentation  `json:"presg"`
}

type BookMove struct {
	Fen        string       `json:"fen"`
	San        string       `json:"san"`
	Annot      string       `json:"annot"`
	Comment    string       `json:"comment"`
	Open       bool         `json:"open"`
	Hasscore   bool         `json:"hasscore"`
	Scorecp    bool         `json:"scorecp"`
	Scoremate  bool         `json:"scoremate"`
	Score      int          `json:"score"`
	Depth      int          `json:"depth"`
}

func (bm *BookMove) checksanity() {
	if (!bm.Scorecp) && (!bm.Scoremate){
		bm.Scorecp=true
	}
}

type BookPosition struct {
	Fen          string                `json:"fen"`
	Moves        map[string]BookMove   `json:"moves"`
	Notes        string                `json:"notes"`
	Arrowalgebs  []string              `json:"arrowalgebs"`
}

func (bp *BookPosition) checksanity() {
	newmoves:=map[string]BookMove{}
	for san , bm := range bp.Moves {
		bm.checksanity()
		newmoves[san]=bm
	}
	bp.Moves=newmoves
}

type Book struct {
	Positions  map[string]BookPosition    `json:"positions"`
	Essay      string                     `json:"essay"`
}

func (bk *Book) checksanity() {
	newpositions:=map[string]BookPosition{}
	for fen , bp := range bk.Positions {
		bp.checksanity()
		newpositions[fen]=bp
	}
	bk.Positions=newpositions
}