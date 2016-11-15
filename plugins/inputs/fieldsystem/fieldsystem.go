package fieldsystem

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"strings"
)

type FieldSystem struct {
	FullRecord bool

	fs         *Fscom
	prevFields map[string]interface{}
}

var FieldSystemConfig = `
  ## Record all fields each gather period, rather
  ## than just differences
  #full_record = false 
`

func (s *FieldSystem) SampleConfig() string {
	return FieldSystemConfig
}

func (s *FieldSystem) Description() string {
	return "Query the Field system state"
}

func (s *FieldSystem) Gather(acc telegraf.Accumulator) (err error) {
	if s.fs == nil {
		s.fs, err = GetSHM()
		if err != nil {
			return err
		}
	}

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	// Ignore if FS not running
	if l, err := s.fs.SemLocked("fs"); !l || err != nil {
		fields["fs"] = false
		acc.AddFields("fs", fields, tags)
		return nil
	}

	// FS semephores
	for i := 0; i < int(s.fs.Sem.Allocated); i++ {
		semname := strings.TrimSpace(string(s.fs.Sem.Name[i][:]))
		semval, err := s.fs.SemLocked(semname)
		if err != nil {
			continue
		}
		fields[semname] = semval
	}

	// FS bools
	fields["data_valid"] = (s.fs.Data_valid[0].User_dv != 0)
	fields["tracking"] = (s.fs.Ionsor != 0)

	// FS strings
	fields["log"] = fsstr(s.fs.LLOG[:])
	fields["schedule"] = fsstr(s.fs.LSKD[:])
	fields["source"] = fsstr(s.fs.Lsorna[:])

	// TODO: diff

	acc.AddFields("fs", fields, tags)
	return nil
}

func fsstr(s []byte) string {
	return strings.TrimSpace(string(s))
}

func init() {
	inputs.Add("fieldsystem", func() telegraf.Input { return &FieldSystem{} })
}
