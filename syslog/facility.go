package syslog

import "fmt"

// A Syslog Priority is a combination of Severity and Facility normally 0 to 191

// RFC5424 - 6.2.1 "The number [.sic.] is known as the Priority value (PRIVAL) and represents both the
// Facility and Severity.  The Priority value consists of one, two, or three decimal integers. [...] The Priority
// value is calculated by first multiplying the Facility number by 8 and then adding the numerical value of the Severity"

// For example, a Priority value of 13 is “user”[1] Facility and “notice”[5] Severity. (1*8)+5=13

type Facility int

var ErrFacility = fmt.Errorf("Not a designated RFC5424 Facility")

// RFC5424 Facilities
// iota represents successive untyped integer constants.
const (
	LogKern Facility = iota
	LogUser
	LogMail
	LogDaemon
	LogAuth
	LogSyslog
	LogLPR
	LogNews
	LogUUCP
	LogCron
	LogAuthPriv
	LogFTP
	LogNTP
	LogAudit
	LogAlert
	LogAt
	LogLocal0
	LogLocal1
	LogLocal2
	LogLocal3
	LogLocal4
	LogLocal5
	LogLocal6
	LogLocal7
)

// Facility Mapping 0 - 23
var facilities = map[string]Facility{
	"kern":     LogKern,
	"user":     LogUser,
	"mail":     LogMail,
	"daemon":   LogDaemon,
	"auth":     LogAuth,
	"syslog":   LogSyslog,
	"lpr":      LogLPR,
	"news":     LogNews,
	"uucp":     LogUUCP,
	"cron":     LogCron,
	"authpriv": LogAuthPriv,
	"ftp":      LogFTP,
	"ntp":      LogNTP,
	"audit":    LogAudit,
	"alert":    LogAlert,
	"at":       LogAt,
	"local0":   LogLocal0,
	"local1":   LogLocal1,
	"local2":   LogLocal2,
	"local3":   LogLocal3,
	"local4":   LogLocal4,
	"local5":   LogLocal5,
	"local6":   LogLocal6,
	"local7":   LogLocal7,
}

// Facility returns the named facility. It returns ErrFacility if the facility
// does not exist.
func FacilityMap(name string) (Facility, error) {
	p, ok := facilities[name]
	if !ok {
		return 0, ErrFacility
	}
	return p, nil
}

func Facilitykeymap(m map[string]Facility, value int) (key string, ok bool) {
	for k, v := range m {
		if int(v) == value {
			key = k
			ok = true
			return
		}
	}
	return
}
