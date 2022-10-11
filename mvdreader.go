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
    X float32 `json:"x"`
    Y float32 `json:"y"`
    Z float32 `json:"z"`
}

func (v *Vector) Set(x, y, z float32) {
	v.X = x
	v.Y = y
	v.Z = z
}

type PE_Info struct {
	Events PE_TYPE `json:"events"`
	Pnum   byte  `json:"player_number"`
}

type Player struct {
	EventInfo   PE_Info `json:"event_info"`
	Name        ReaderString `json:"name"`
	Team        ReaderString `json:"team"`
	Userid      int `json:"user_id"`
	Spectator   bool `json:"spectator"`
	Deaths      int `json:"deaths"`
	Suicides    int `json:"suicides"`
	Teamkills   int `json:"teamkills"`
	Origin      Vector `json:"origin"`
	Angle       Vector `json:"angle"`
	ModelIndex  byte `json:"model_index"`
	SkinNum     byte `json:"skin_number"`
	WeaponFrame byte `json:"weapon_frame"`
	Effects     byte `json:"effects"`
	Ping        int `json:"ping"`
	Pl          byte `json:"packetloss"`
	Entertime   float32 `json:"enter_time"`
	Frame       int `json:"frame"`

	// stat
	Health       int `json:"health"`
	Frags        int `json:"frags"`
	Weapon       int `json:"weapon"`
	Ammo         int `json:"ammo"`
	Armor        int `json:"armor"`
	Weaponframe  int `json:"weapon_frame"`
	Shells       int `json:"shell"`
	Nails        int `json:"nails"`
	Rockets      int `json:"rockets"`
	Cells        int `json:"cells"`
	Activeweapon int `json:"active_weapon"`
	Totalsecrets int `json:"total_secrets"`
	Totalmonster int `json:"total_monsters"`
	Secrets      int `json:"secrets"`
	Monsters     int `json:"monsters"`
	Items        int `json:"items"`
	Viewheight   int `json:"viewheight"`
	Time         int `json:"time"`

	Setinfo map[string]ReaderString `json:"setinfo"`
}

func (p *Player) HasItem(item IT_TYPE) bool {
	return p.Items&int(item) == int(item)
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
	Key   ReaderString
    Value ReaderString
}

type ServerMessage struct {
	From    int
	Message ReaderString
}

type State struct {
	Time         float64
	Players      [32]Player
	SoundsActive []Sound
	SoundsStatic []Sound
	Serverinfo   []Serverinfo
	Messages     []ServerMessage
	Centerprint  []ReaderString
	TempEntities []Tempentity
	Entities     []Entity
	StuffText    []ReaderString
    ProtocolMessage []SVC_TYPE
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
	Gamedir         ReaderString
	Demotime        float32
	Mapname         ReaderString 
	Hostname        ReaderString
	Movevars        []float32
	Serverinfo      map[string]ReaderString
	Soundlist       []ReaderString
	Modellist       []ReaderString
	Baseline        []*Entity
	baselineIndexed map[int]*Entity
	StaticEntities  []Entity
	Paused          bool
}

type Mvd struct {
	debug *log.Logger

    ascii_table []rune

	file        []byte
	file_offset uint
	file_length uint
	Frame       uint
	done        bool
	trace       traceInfo

	demo             Demo
	Server           Server
	State            State
	State_last_frame State
}

func Load(input []byte, logger *log.Logger, ascii_table *string) (Mvd, error) {
	var mvd Mvd
	mvd.debug = logger
	mvd.file = input
	mvd.file_length = uint(len(input))
	mvd.Server.Serverinfo = make(map[string]ReaderString)
	mvd.Server.baselineIndexed = make(map[int]*Entity)
	// FIXME - this is garbage we need to fix this in some other way
    var ml []ReaderString
	mvd.Server.Modellist = ml
    var sl []ReaderString
	mvd.Server.Soundlist = sl
	mvd.trace.enabled = false
    mvd.ascii_table = AsciiTableInit(ascii_table)
	return mvd, nil
}

func (mvd *Mvd) Parse() error {
	for {
		done, err := mvd.ParseFrame()
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

func (mvd *Mvd) ParseFrame() ( bool, error) {
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
	mvd.State_last_frame.ProtocolMessage = mvd.State.ProtocolMessage
	mvd.State.ProtocolMessage = nil
    mvd.State.StuffText = nil
	for i := range mvd.State.Players {
		mvd.State.Players[i].Setinfo = make(map[string]ReaderString)
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
				return mvd.done, nil
			}
			return false, err
		}

        if mvd.done {
            mvd.traceParseFrameStop()
            return mvd.done, nil
        }
		mvd.Frame++
		readahead_time, err := mvd.demotimeReadahead()
        if err != nil {
			return false, err
        }

		if readahead_time != 0 {
			break
		}
	}

	mvd.traceParseFrameStop()
	return mvd.done, nil
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
					i, err := mvd.readInt()
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
			outgoing_sequence, err := mvd.readUint()
			if err != nil {
				return err
			}

			mvd.traceReadTraceAdditionalInfo("outgoing_sequence", outgoing_sequence)
			mvd.demo.outgoing_sequence = outgoing_sequence

			mvd.traceAddReadTrace("incoming_sequence")
			incoming_sequence, err := mvd.readUint()
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
			b, err := mvd.readIt(msg_type)
			if err != nil {
				return err
			}
			for b {
				if mvd.debug != nil {
					mvd.debug.Println("did we loop?")
				}
				b, err = mvd.readIt(msg_type)
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
