package mvdreader

import (
	"errors"
	"log"
	"strconv"
)

type Demo struct {
	time                                 float64
	last_to                              uint
	last_type                            DEM_TYPE
	outgoing_sequence, incoming_sequence uint32
	protocol                             PROTOCOL_VERSION
	fte_pext                             FTE_PROTOCOL_EXTENSION
	fte_pext2                            FTE_PROTOCOL_EXTENSION
	mvd_pext                             MVD_PROTOCOL_EXTENSION
}

type Vector struct {
	X, Y, Z float32
}

func (v *Vector) Set(x, y, z float32) {
	v.X = x
	v.Y = y
	v.Z = z
}

type PE_Info struct {
	Events PE_TYPE
	Pnum   byte
}

type Player struct {
	EventInfo   PE_Info
	Name        string
	Team        string
	Userid      int
	Spectator   bool
	Deaths      int
	Suicides    int
	Teamkills   int
	Origin      Vector
	Angle       Vector
	ModelIndex  byte
	SkinNum     byte
	WeaponFrame byte
	Effects     byte
	Ping        int
	Pl          byte
	Entertime   float32
	Frame       int

	// stat
	Health       int
	Frags        int
	Weapon       int
	Ammo         int
	Armor        int
	Weaponframe  int
	Shells       int
	Nails        int
	Rockets      int
	Cells        int
	Activeweapon int
	Totalsecrets int
	Totalmonster int
	Secrets      int
	Monsters     int
	Items        int
	Viewheight   int
	Time         int

	Setinfo map[string]string
}

type Sound struct {
	Frame       uint
	Index       byte
	Channel     SND_TYPE
	Volume      byte
	Attenuation byte
	Origin      Vector
}

type Serverinfo struct {
	Key   string
	Value string
}

type ServerMessage struct {
	From    int
	Message string
}

type State struct {
	Time         float64
	Players      [32]Player
	SoundsActive []Sound
	SoundsStatic []Sound
	Serverinfo   []Serverinfo
	Messages     []ServerMessage
	Centerprint  []string
	TempEntities []Tempentity
	Entities     []Entity
	StuffText    []string
}

type Entity struct {
	Index         int
	ModelIndex    byte
	SkinNum       byte
	Frame         byte
	ColorMap      byte
	Effects       byte
	Origin, Angle Vector
	Transparency  byte
}

type Tempentity struct {
	Type       TE_TYPE
	ModelIndex int
	SkinNum    int
	Frame      int
	ColorMap   int
	Origin     Vector
	Start      Vector // beams
	Entity     int    // beams?
	Count      int    // gunshot, blood
}

type Server struct {
	ServerCount     int
	Gamedir         string
	Demotime        float32
	Mapname         string
	Hostname        string
	Movevars        []float32
	Serverinfo      map[string]string
	Soundlist       []string
	Modellist       []string
	Baseline        []*Entity
	baselineIndexed map[int]*Entity
	StaticEntities  []Entity
	Paused          bool
}

type Mvd struct {
	debug *log.Logger

	file        []byte
	file_offset uint
	file_length uint
	filename    string
	Frame       uint
	done        bool
	trace       traceInfo

	demo             Demo
	Server           Server
	State            State
	State_last_frame State
}

func Load(input []byte, logger *log.Logger) (error, Mvd) {
	var mvd Mvd
	mvd.debug = logger
	mvd.file = input
	mvd.file_length = uint(len(input))
	mvd.Server.Serverinfo = make(map[string]string, 0)
	mvd.Server.baselineIndexed = make(map[int]*Entity, 0)
	// FIXME - this is garbage we need to fix this in some other way
	mvd.Server.Modellist = append(mvd.Server.Modellist, "")
	mvd.trace.enabled = false
	return nil, mvd
}

func (mvd *Mvd) Parse() error {
	for {
		err, done := mvd.ParseFrame()
		if err != nil {
			return err
		}
		if done {
			break
		}
	}
	return nil
}

func (mvd *Mvd) TraceEnable() {
	mvd.trace.enabled = true
}

func (mvd *Mvd) TraceGet() []*TraceParseFrame {
	return mvd.trace.traces
}

func (mvd *Mvd) ParseFrame() (error, bool) {
	if mvd.debug != nil {
		mvd.debug.Printf("Frame (%v)", mvd.Frame)
	}
	mvd.State_last_frame = State(mvd.State)
	mvd.State_last_frame.Serverinfo = mvd.State.Serverinfo
	mvd.State_last_frame.TempEntities = mvd.State.TempEntities
	mvd.State.TempEntities = nil
	mvd.State.Centerprint = nil
	mvd.State_last_frame.SoundsActive = mvd.State.SoundsActive
	mvd.State.SoundsActive = nil
	mvd.State.Serverinfo = nil
	mvd.State_last_frame.Messages = mvd.State.Messages
	mvd.State.Messages = nil
	for i := range mvd.State.Players {
		mvd.State.Players[i].Setinfo = make(map[string]string)
		for k, v := range mvd.State_last_frame.Players[i].Setinfo {
			mvd.State.Players[i].Setinfo[k] = v
		}
	}

	mvd.traceParseFrameStart()

	for {
		err := mvd.readFrame()
		if err != nil {
			if mvd.done {
				mvd.traceParseFrameStop()
				return nil, mvd.done
			}
			return err, false
		}
		mvd.Frame++
		err, readahead_time := mvd.demotimeReadahead()

		if readahead_time != 0 {
			break
		}
	}

	mvd.traceParseFrameStop()
	return nil, mvd.done
}

func (mvd *Mvd) readFrame() error {
	for {
		mvd.traceReadFrameStart()
		mvd.demotime()
		mvd.State.Time = mvd.demo.time

		mvd.traceAddReadTrace("cmd")
		err, cmd := mvd.readByte()
		if err != nil {
			return err
		}
		msg_type := DEM_TYPE(cmd & 7)
		mvd.traceReadTraceAdditionalInfo("mvd_type", msg_type)
		if msg_type == dem_cmd {
			return errors.New("this is an mvd parser")
		}

		if mvd.debug != nil {
			mvd.debug.Println("handling cmd", DEM_TYPE(cmd))
		}
		if msg_type >= dem_multiple && msg_type <= dem_all {
			switch msg_type {
			case dem_multiple:
				{
					mvd.traceAddReadTrace("dem_multiple")
					err, i := mvd.readInt()
					if err != nil {
						return err
					}
					mvd.demo.last_to = uint(i)
					mvd.traceReadTraceAdditionalInfo("last_to", uint(i))
					mvd.traceReadTraceAdditionalInfo("last_to_binary", strconv.FormatInt(int64(mvd.demo.last_to), 2))
					if mvd.debug != nil {
						mvd.debug.Println(
							"affected players: ",
							strconv.FormatInt(int64(mvd.demo.last_to), 2),
							mvd.demo.last_to,
						)
					}
					mvd.demo.last_type = dem_multiple
					break
				}
			case dem_single:
				{
					mvd.demo.last_to = uint(cmd >> 3)
					mvd.demo.last_type = dem_single
					break
				}
			case dem_all:
				{
					if mvd.debug != nil {
						mvd.debug.Println("dem_all", mvd.file_offset)
					}
					mvd.demo.last_to = 0
					mvd.demo.last_type = dem_all
					break
				}

			case dem_stats:
				{
					if mvd.debug != nil {
						mvd.debug.Println("dem_all", mvd.file_offset)
						mvd.debug.Println(
							"dem_stats",
							cmd,
							cmd&7,
							dem_stats,
							mvd.file_offset,
							"byte: ",
							mvd.file[mvd.file_offset],
						)
					}
					mvd.traceReadTraceAdditionalInfo("last_to", uint(cmd>>3))
					mvd.demo.last_to = uint(cmd >> 3)
					mvd.demo.last_type = dem_stats
					break
				}
			}
			msg_type = dem_read
		}
		if msg_type == dem_set {
			if mvd.debug != nil {
				mvd.debug.Println("dem_set", mvd.file_offset)
			}

			mvd.traceAddReadTrace("outgoing_sequence")
			err, outgoing_sequence := mvd.readUint()
			if err != nil {
				return err
			}

			mvd.traceReadTraceAdditionalInfo("outgoing_sequence", outgoing_sequence)
			mvd.demo.outgoing_sequence = outgoing_sequence

			mvd.traceAddReadTrace("incoming_sequence")
			err, incoming_sequence := mvd.readUint()
			if err != nil {
				return err
			}

			mvd.demo.incoming_sequence = incoming_sequence
			if mvd.debug != nil {
				mvd.debug.Printf("Squence in(%v) out(%v)", mvd.demo.incoming_sequence, mvd.demo.outgoing_sequence)
			}
			mvd.traceReadFrameStop()
			continue
		}
		if msg_type == dem_read {
			err, b := mvd.readIt(msg_type)
			if err != nil {
				return err
			}
			for b {
				if mvd.debug != nil {
					mvd.debug.Println("did we loop?")
				}
				err, b = mvd.readIt(msg_type)
				if err != nil {
					return err
				}
			}
			mvd.traceReadFrameStop()
			return nil
		}
		if mvd.debug != nil {
			mvd.debug.Println(cmd)
		}

		mvd.traceReadFrameStop()
		return errors.New("this should not happen")
	}
}
