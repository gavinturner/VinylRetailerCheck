package scripts

import (
	"fmt"
	"github.com/gavinturner/vinylretailers/cmd"
	"github.com/gavinturner/vinylretailers/db"
	"strings"
)

func convertImageStrings() error {

	psqlDB, err := cmd.InitialiseDbConnection()
	if err != nil {
		return err
	}
	defer psqlDB.Close()
	vinylDS := db.NewDB(psqlDB)

	skusA, err := vinylDS.GetAllSKUs(nil)
	if err != nil {
		return err
	}
	for i := 0; i < len(skusA); i++ {
		s := skusA[i]
		if strings.Index(s.ImageUrl, "<img width=\"150px\" height=\"150px\" src=\"") == 0 {
			s.ImageUrl = s.ImageUrl[len("<img width=\"150px\" height=\"150px\" src=\""):]
		}
		if strings.Index(s.ImageUrl, "\" />") > 0 {
			s.ImageUrl = s.ImageUrl[0:strings.Index(s.ImageUrl, "\" />")]
		}
		if strings.Index(s.ImageUrl, "\"/>") > 0 {
			s.ImageUrl = s.ImageUrl[0:strings.Index(s.ImageUrl, "\"/>")]
		}
		if strings.Index(s.ImageUrl, "\">") > 0 {
			s.ImageUrl = s.ImageUrl[0:strings.Index(s.ImageUrl, "\">")]
		}
		if strings.Index(s.ImageUrl, "\" >") > 0 {
			s.ImageUrl = s.ImageUrl[0:strings.Index(s.ImageUrl, "\" >")]
		}
		skusA[i] = s
	}

	for _, s := range skusA {
		fmt.Printf("%v\n", s.ImageUrl)
		err = vinylDS.UpdateSKU(nil, &s)
		if err != nil {
			return err
		}
	}
	return nil
}
