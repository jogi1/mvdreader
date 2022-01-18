package mvdreader

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Message struct {
	size        uint
	data        []byte
	offset      uint
	OffsetStart uint
	mvd         *Mvd
}

func (mvd *Mvd) emitEventPlayer(player *Player, pnum byte, pe_type PE_TYPE) {
	player.EventInfo.Pnum = pnum
	player.EventInfo.Events |= pe_type
}

func (mvd *Mvd) emitEventSound(sound *Sound) {
	sound.Frame = mvd.Frame
	mvd.State.SoundsActive = append(mvd.State.SoundsActive, *sound)
}

func (mvd *Mvd) messageParse(message Message) (error, bool) {
	message.mvd = mvd
	for {
		if mvd.done == true {
			return nil, true
		}

		mvd.traceStartMessageTrace(&message)
		mvd.traceStartMessageTraceReadTrace("type")

		err, msgt := message.readByte()
		if err != nil {
			return err, false
		}
		msg_type := SVC_TYPE(msgt)
		mvd.traceMessageInfo(msg_type)

		if mvd.debug != nil {
			mvd.debug.Println("handling: ", msg_type)
			mvd.debug.Println("expected function: ", strings.Title(fmt.Sprintf("%s", msg_type)))
		}
		m := reflect.ValueOf(&message).MethodByName(strings.Title(fmt.Sprintf("%s", msg_type)))

		if m.IsValid() {
			m.Call([]reflect.Value{reflect.ValueOf(mvd)})
		} else {
			return errors.New(fmt.Sprintf("error for message type: %#v %#v", msg_type, m)), false
		}
		if message.offset >= message.size {
			return nil, true
		}
		if mvd.done {
			return nil, true
		}
	}
	if message.offset != message.size {
		return errors.New(fmt.Sprint("did not read message fully ", message.offset, message.size)), false
	}
	return nil, true
}

func (message *Message) Svc_serverdata(mvd *Mvd) error {
	var mrt *TraceRead
	for {
		message.traceAddMessageReadTrace("protocol")
		err, prot := message.readLong()
		if err != nil {
			return err
		}
		if mrt != nil {
			mrt.Value = prot
		}
		message.mvd.demo.protocol = PROTOCOL_VERSION(prot)
		protocol := message.mvd.demo.protocol

		if protocol == protocol_fte2 {
			message.traceAddMessageReadTrace("protocol_fte2")
			err, fte_pext2 := message.readLong()
			if err != nil {
				return err
			}
			message.mvd.demo.fte_pext2 = FTE_PROTOCOL_EXTENSION(fte_pext2)
			continue
		} else if protocol == protocol_fte {
			message.traceAddMessageReadTrace("protocol_fte")
			err, fte_pext := message.readLong()
			if err != nil {
				return err
			}
			message.mvd.demo.fte_pext = FTE_PROTOCOL_EXTENSION(fte_pext)
			continue
		} else if protocol == protocol_mvd1 {
			message.traceAddMessageReadTrace("protocol_mvd")
			err, mvd_pext := message.readLong()
			if err != nil {
				return err
			}
			message.mvd.demo.mvd_pext = MVD_PROTOCOL_EXTENSION(mvd_pext)
			continue
		} else if protocol == protocol_standard {
			break
		} else {
			return fmt.Errorf("protocol broke!: %d", protocol)
		}
	}

	message.traceAddMessageReadTrace("server_count")
	err, server_count := message.readLong() // server count
	if err != nil {
		return err
	}
	mvd.Server.ServerCount = server_count

	message.traceAddMessageReadTrace("gamedir")
	err, gamedir := message.readString() // gamedir
	if err != nil {
		return err
	}
	mvd.Server.Gamedir = gamedir

	message.traceAddMessageReadTrace("demotime")
	err, demotime := message.readFloat() // demotime
	if err != nil {
		fmt.Println(err)
		return err
	}
	mvd.Server.Demotime = demotime

	message.traceAddMessageReadTrace("map")
	err, s := message.readString()
	if err != nil {
		return err
	}
	mvd.Server.Mapname = s
	for i := 0; i < 10; i++ {

		message.traceAddMessageReadTrace(fmt.Sprintf("movevar_%d", i))
		err, mv := message.readFloat()
		if err != nil {
			fmt.Println(err)
			return err
		}
		mvd.Server.Movevars = append(mvd.Server.Movevars, mv)
	}
	return nil
}

/*
func (message *Message) Svc_bad(mvd *Mvd) {
}
*/

func (message *Message) Svc_cdtrack(mvd *Mvd) error {
	message.traceAddMessageReadTrace("track")
	err, _ := message.readByte()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_stufftext(mvd *Mvd) error {
	message.traceAddMessageReadTrace("stufftext")
	err, s := message.readString()
	if err != nil {
		return err
	}

	message.mvd.State.StuffText = append(message.mvd.State.StuffText, s)

	if strings.HasPrefix(s, "fullserverinfo") {
		trim := s[len("fullserverinfo \"\\"):]
		trim = strings.TrimRight(trim, "\\\"")
		splits := strings.Split(trim, "\\")

		for i := 0; i < len(splits); i += 2 {
			message.mvd.Server.Serverinfo[splits[i]] = splits[i+1]
		}
	}
	return nil
}

func (message *Message) Svc_soundlist(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index_start")
	err, _ := message.readByte() // those are some indexes
	if err != nil {
		return err
	}
	for {
		message.traceAddMessageReadTrace("name")
		err, s := message.readString()
		if err != nil {
			return err
		}
		if len(s) == 0 {
			break
		}
		message.mvd.Server.Soundlist = append(message.mvd.Server.Soundlist, s)
	}

	message.traceAddMessageReadTrace("offset")
	err, _ = message.readByte() // some more indexes

	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_modellist(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index_start")
	err, _ := message.readByte() // those are some indexes
	if err != nil {
		return err
	}
	for {
		message.traceAddMessageReadTrace("name")
		err, s := message.readString()
		if err != nil {
			return err
		}
		if len(s) == 0 {
			break
		}
		message.mvd.Server.Modellist = append(message.mvd.Server.Modellist, s)
	}
	message.traceAddMessageReadTrace("offset")
	err, _ = message.readByte() // some more indexes
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_spawnbaseline(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index")
	err, index := message.readShort()
	if err != nil {
		return err
	}
	entity, err := message.parseBaseline(mvd)
	entity.Index = index
	mvd.Server.baselineIndexed[index] = entity
	if err != nil {
		return err
	}

	mvd.Server.Baseline = append(mvd.Server.Baseline, entity)
	return nil
}

func (message *Message) Svc_updatefrags(mvd *Mvd) error {
	message.traceAddMessageReadTrace("pnum")
	err, player := message.readByte()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("frags")
	err, frags := message.readShort()
	if err != nil {
		return err
	}
	mvd.State.Players[int(player)].Frags = int(frags)
	mvd.emitEventPlayer(&mvd.State.Players[int(player)], player, PE_STATS)
	return nil
}

func (message *Message) Svc_playerinfo(mvd *Mvd) error {
	var pe_type PE_TYPE
	message.traceAddMessageReadTrace("pnum")
	err, pnum := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]

	message.traceAddMessageReadTrace("flags")
	err, sflags := message.readShort()
	if err != nil {
		return err
	}
	flags := DF_TYPE(sflags)

	message.traceAddMessageReadTrace("frame")
	err, frame := message.readByte()
	p.Frame = int(frame)
	if err != nil {
		return err
	}
	mvdPrint("frame: ", frame)
	for i := 0; i < 3; i++ {
		t := DF_ORIGIN << i
		if flags&t == t {
			pe_type |= PE_MOVEMENT
			flags -= t
			switch i {
			case 0:
				{
					message.traceAddMessageReadTrace("Origin.X")
					err, coord := message.readCoord()
					if err != nil {
						return err
					}
					p.Origin.X = coord
				}
			case 1:
				{
					message.traceAddMessageReadTrace("Origin.Y")
					err, coord := message.readCoord()
					if err != nil {
						return err
					}
					p.Origin.Y = coord
				}
			case 2:
				{
					message.traceAddMessageReadTrace("Origin.Z")
					err, coord := message.readCoord()
					if err != nil {
						return err
					}
					p.Origin.Z = coord
				}
			}
		}
	}
	for i := 0; i < 3; i++ {
		t := DF_ANGLES << i
		if flags&t == t {
			pe_type |= PE_MOVEMENT
			flags -= t
			switch i {
			case 0:
				{
					message.traceAddMessageReadTrace("Angle.X")
					err, angle := message.readAngle16()
					if err != nil {
						return err
					}
					p.Angle.X = angle
				}
			case 1:
				{
					message.traceAddMessageReadTrace("Angle.Y")
					err, angle := message.readAngle16()
					if err != nil {
						return err
					}
					p.Angle.Y = angle
				}
			case 2:
				{
					message.traceAddMessageReadTrace("Angle.Z")
					err, angle := message.readAngle16()
					if err != nil {
						return err
					}
					p.Angle.Z = angle
				}
			}
		}
	}

	mvdPrint(flags)

	if flags&DF_MODEL == DF_MODEL {
		pe_type |= PE_ANIMATION

		message.traceAddMessageReadTrace("model")
		err, mindex := message.readByte()
		if err != nil {
			return err
		}
		p.ModelIndex = mindex // modelindex
	}

	if flags&DF_SKINNUM == DF_SKINNUM {
		pe_type |= PE_ANIMATION
		message.traceAddMessageReadTrace("skinnum")
		err, skinnum := message.readByte()
		if err != nil {
			return err
		}
		p.SkinNum = skinnum // skinnum
	}

	if flags&DF_EFFECTS == DF_EFFECTS {
		pe_type |= PE_ANIMATION
		message.traceAddMessageReadTrace("effects")
		err, effects := message.readByte()
		if err != nil {
			return err
		}
		p.Effects = effects // effects
	}

	if flags&DF_WEAPONFRAME == DF_WEAPONFRAME {
		pe_type |= PE_ANIMATION

		message.traceAddMessageReadTrace("weaponframe")
		err, weaponframe := message.readByte()
		if err != nil {
			return err
		}
		p.WeaponFrame = weaponframe // weaponframe
	}
	if flags != 0 {
		return fmt.Errorf("svc_player: flags not fully parsed")
	}

	mvd.emitEventPlayer(p, pnum, pe_type)
	return nil
}

func (message *Message) Svc_updateping(mvd *Mvd) error {
	message.traceAddMessageReadTrace("pnum")
	err, pnum := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]

	message.traceAddMessageReadTrace("ping")
	err, ping := message.readShort()
	if err != nil {
		return err
	}
	p.Ping = ping // ping
	mvd.emitEventPlayer(p, pnum, PE_NETWORKINFO)
	return nil
}

func (message *Message) Svc_updatepl(mvd *Mvd) error {
	message.traceAddMessageReadTrace("pnum")
	err, pnum := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]

	message.traceAddMessageReadTrace("pl")
	err, pl := message.readByte()
	if err != nil {
		return err
	}
	p.Pl = pl // pl
	mvd.emitEventPlayer(p, pnum, PE_NETWORKINFO)
	return nil
}

func (message *Message) Svc_updateentertime(mvd *Mvd) error {
	message.traceAddMessageReadTrace("pnum")

	err, pnum := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]

	message.traceAddMessageReadTrace("entertime")
	err, entertime := message.readFloat()
	if err != nil {
		return err
	}
	p.Entertime = entertime // entertime
	mvd.emitEventPlayer(p, pnum, PE_NETWORKINFO)
	return nil
}

func (message *Message) Svc_updateuserinfo(mvd *Mvd) error {
	message.traceAddMessageReadTrace("pnum")
	err, pnum := message.readByte()
	if err != nil {
		return err
	}

	message.traceAddMessageReadTrace("uid")
	err, uid := message.readLong()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]
	p.Userid = uid

	message.traceAddMessageReadTrace("userinfo")
	err, ui := message.readString()
	if err != nil {
		return err
	}
	if len(ui) < 2 {
		return nil
	}
	ui = ui[1:]
	splits := strings.Split(ui, "\\")

	p.Spectator = false
	p.Setinfo["*spectator"] = "0"
	for i := 0; i < len(splits); i += 2 {
		v := splits[i+1]
		switch splits[i] {
		case "name":
			p.Name = v
		case "team":
			p.Team = v
		case "*spectator":
			if v == "1" {
				p.Spectator = true
			}
		}
		p.Setinfo[splits[i]] = v
	}
	mvd.emitEventPlayer(p, pnum, PE_USERINFO)
	return nil
}

func (message *Message) Svc_sound(mvd *Mvd) error {
	var s Sound
	message.traceAddMessageReadTrace("channel")
	err, sc := message.readShort()
	if err != nil {
		return err
	}
	channel := SND_TYPE(sc) // channel
	s.Channel = channel
	if channel&SND_VOLUME == SND_VOLUME {
		message.traceAddMessageReadTrace("volume")
		err, volume := message.readByte()
		if err != nil {
			return err
		}
		s.Volume = volume
	}

	if channel&SND_ATTENUATION == SND_ATTENUATION {
		message.traceAddMessageReadTrace("attenuation")
		err, attenuation := message.readByte()
		if err != nil {
			return err
		}
		s.Attenuation = attenuation
	}
	ent := (s.Channel >> 3) & 1023
	message.traceMessageReadAdditionlInfo("ent", ent)
	message.traceMessageReadAdditionlInfo("actual channel", s.Channel&7)
	message.traceAddMessageReadTrace("index")
	err, index := message.readByte()
	if err != nil {
		return err
	}
	s.Index = index // sound_num

	message.traceAddMessageReadTrace("Origin.X")
	err, x := message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Y")
	err, y := message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Z")
	err, z := message.readCoord()
	if err != nil {
		return err
	}
	s.Origin.Set(x, y, z)
	mvd.emitEventSound(&s)
	return nil
}

func (message *Message) Svc_spawnstaticsound(mvd *Mvd) error {
	var s Sound
	message.traceAddMessageReadTrace("Origin.X")
	err, x := message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Y")
	err, y := message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Z")
	err, z := message.readCoord()
	if err != nil {
		return err
	}
	s.Origin.Set(x, y, z)

	message.traceAddMessageReadTrace("index")
	err, index := message.readByte()
	if err != nil {
		return err
	}
	s.Index = index // sound_num

	message.traceAddMessageReadTrace("volume")
	err, volume := message.readByte()
	if err != nil {
		return err
	}
	s.Volume = volume // sound volume
	message.traceAddMessageReadTrace("attenuation")
	err, attenuation := message.readByte()
	if err != nil {
		return err
	}
	s.Attenuation = attenuation // sound attenuation
	mvd.State.SoundsStatic = append(mvd.State.SoundsStatic, s)
	return nil
}

func (message *Message) Svc_setangle(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index")
	err, _ := message.readByte() // something weird?
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.X")
	err, _ = message.readAngle() // x
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.Y")
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.Z")
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_lightstyle(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index")
	err, _ := message.readByte() // lightstyle num
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("style")
	err, _ = message.readString()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_updatestatlong(mvd *Mvd) error {
	message.traceAddMessageReadTrace("stat")
	err, b := message.readByte()
	if err != nil {
		return err
	}
	stat := STAT_TYPE(b)
	message.traceAddMessageReadTrace("value")
	err, value := message.readLong()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[mvd.demo.last_to]
	s := fmt.Sprintf("%s", STAT_TYPE(stat))
	s = strings.TrimPrefix(s, "STAT_")
	s = strings.ToLower(s)
	s = strings.Title(s)
	ps := reflect.ValueOf(p)
	st := ps.Elem()
	f := st.FieldByName(s)
	if f.IsValid() {
		if f.CanSet() {
			if f.Kind() == reflect.Int {
				f.SetInt(int64(value))
			}
		}
	}
	mvd.emitEventPlayer(p, byte(mvd.demo.last_to), PE_STATS)
	return nil
}

func (message *Message) Svc_updatestat(mvd *Mvd) error {
	message.traceAddMessageReadTrace("stat")
	err, b := message.readByte()
	if err != nil {
		return err
	}
	stat := STAT_TYPE(b)

	message.traceAddMessageReadTrace("value")
	err, value := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[mvd.demo.last_to]
	s := STAT_TYPE(stat).String()
	s = strings.TrimPrefix(s, "STAT_")
	s = strings.ToLower(s)
	s = strings.Title(s)
	ps := reflect.ValueOf(p)
	st := ps.Elem()
	f := st.FieldByName(s)
	if f.IsValid() {
		if f.CanSet() {
			if f.Kind() == reflect.Int {
				f.SetInt(int64(value))
			}
		}
	} else {
		return fmt.Errorf("unknown STAT_ type: %s", stat)
	}
	mvd.emitEventPlayer(p, byte(mvd.demo.last_to), PE_STATS)
	return nil
}

func (message *Message) ParseDelta(mvd *Mvd, from, to *Entity) error {
	return nil
}

func (message *Message) Svc_deltapacketentities(mvd *Mvd) error {
	message.traceAddMessageReadTrace("from")
	err, _ := message.readByte()
	if err != nil {
		return err
	}

	for {
		message.traceAddMessageReadTrace("flags")
		err, w := message.readShort()
		if err != nil {
			return err
		}
		if w == 0 {
			break
		}

		message.traceMessageReadAdditionlInfo("num", w&511)
		entNum := w & 511
		var entity *Entity
		for _, e := range mvd.State.Entities {
			if e.Index == entNum {
				entity = &e
				break
			}
		}

		if entity == nil {
			message.traceMessageReadAdditionlInfo("entity", fmt.Sprintf("%d not found", entNum))
			entity = new(Entity)
		}
		w &= ^511
		bits := w

		message.traceMessageReadAdditionlInfo("bits", bits)
		message.traceMessageReadAdditionlInfo("whats the return?", bits&U_MOREBITS == U_MOREBITS)
		message.traceMessageReadAdditionlInfo("morebits", U_MOREBITS)
		if bits&U_MOREBITS == U_MOREBITS {
			message.traceAddMessageReadTrace("morebits")
			err, i := message.readByte()
			if err != nil {
				return err
			}
			bits |= int(i)
		}

		morebits := 0
		if bits&U_FTE_EVENVENMORE == U_FTE_EVENVENMORE {
			err, i := message.readByte()
			if err != nil {
				return err
			}
			morebits = int(i)
			if morebits&U_FTE_YETMORE == U_FTE_YETMORE {
				err, mi := message.readByte()
				if err != nil {
					return err
				}
				morebits = morebits | int(mi)<<8

			}
		}

		if bits&U_MOREBITS == U_MOREBITS {
			if mvd.demo.fte_pext&FTE_PEXT_ENTITYDBL == FTE_PEXT_ENTITYDBL {
				err, evenMore := message.peekByte(0)
				if err != nil {
					return err
				}
				if evenMore&U_FTE_EVENVENMORE == U_FTE_EVENVENMORE {
					err, evenMore = message.peekByte(1)
					if err != nil {
						return err
					}
					if evenMore&U_FTE_ENTITYDBL == U_FTE_ENTITYDBL {
						entNum += 512
					}
					if evenMore&U_FTE_ENTITYDBL2 == U_FTE_ENTITYDBL2 {
						entNum += 1024
					}
				}
			}
		}

		if bits&U_REMOVE == U_REMOVE {
			message.traceMessageReadAdditionlInfo("entity removed", "")
			if mvd.demo.fte_pext&FTE_PEXT_ENTITYDBL == FTE_PEXT_ENTITYDBL && bits&U_MOREBITS == U_MOREBITS {
				message.traceAddMessageReadTrace("fte extension")
				err, ftext := message.readByte()
				if err != nil {
					return err
				}
				if ftext&U_FTE_EVENVENMORE == U_FTE_EVENVENMORE {
					err, _ := message.readByte()
					if err != nil {
						return err
					}
				}
			}
		}
		if bits&U_MODEL == U_MODEL {
			message.traceAddMessageReadTrace("model")
			err, entity.ModelIndex = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_FRAME == U_FRAME {
			message.traceAddMessageReadTrace("frame")
			err, entity.Frame = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_COLORMAP == U_COLORMAP {
			message.traceAddMessageReadTrace("colormap")
			err, entity.ColorMap = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_SKIN == U_SKIN {
			message.traceAddMessageReadTrace("skin")
			err, entity.SkinNum = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_EFFECTS == U_EFFECTS {
			message.traceAddMessageReadTrace("effects")
			err, entity.Effects = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN1 == U_ORIGIN1 {
			message.traceAddMessageReadTrace("Origin.X")
			err, entity.Origin.X = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE1 == U_ANGLE1 {
			message.traceAddMessageReadTrace("Angle.X")
			err, entity.Angle.X = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN2 == U_ORIGIN2 {
			message.traceAddMessageReadTrace("Origin.Y")
			err, entity.Origin.Y = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE2 == U_ANGLE2 {
			message.traceAddMessageReadTrace("Angle.Y")
			err, entity.Angle.Y = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN3 == U_ORIGIN3 {
			message.traceAddMessageReadTrace("Origin.Z")
			err, entity.Origin.Z = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE3 == U_ANGLE3 {
			message.traceAddMessageReadTrace("Angle.Z")
			err, entity.Angle.Z = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_REMOVE == U_REMOVE {
			found := false
			i := -1
			for x, e := range mvd.State.Entities {
				if e.Index == entNum {
					found = true
					i = x
					break
				}
			}
			if found {
				mvd.State.Entities = append(mvd.State.Entities[:i], mvd.State.Entities[i+1:]...)
			} else {
				message.traceMessageReadAdditionlInfo("DEBUG FIND", fmt.Sprintf("index(%d)", entNum))
			}
		}

		// FIXME: do all the other fte stuff
		if morebits&U_FTE_TRANS == U_FTE_TRANS {
			err, b := message.readByte()
			if err != nil {
				return err
			}
			entity.Transparency = b
		}
	}
	return nil
}

func (message *Message) Svc_packetentities(mvd *Mvd) error {
	count := 0
	mvd.State.Entities = nil
	for {
		message.traceMessageReadAdditionlInfo("entity start:", count)
		count++
		message.traceAddMessageReadTrace("bits")
		err, w := message.readShort()
		if err != nil {
			return err
		}
		if w == 0 {
			message.traceMessageReadAdditionlInfo("w", w)
			break
		}

		// lower 8 bits are the entity number
		// upper 8 are the bits
		newnum := w
		newnum &= 511
		w &= ^511
		message.traceMessageReadAdditionlInfo("newnum", newnum)
		entity := new(Entity)
		et := mvd.Server.baselineIndexed[newnum]
		if et != nil {
			*entity = *et
			message.traceMessageReadAdditionlInfo(
				"baseline ",
				fmt.Sprintf("found (%s)", mvd.Server.Modellist[entity.ModelIndex]),
			)
		} else {
			message.traceMessageReadAdditionlInfo("baseline ", fmt.Sprintf("not found index(%d)", newnum))
		}
		entity.Index = newnum
		bits := w

		if bits&U_MOREBITS == U_MOREBITS {
			message.traceAddMessageReadTrace("morebits")
			err, i := message.readByte()
			if err != nil {
				return err
			}
			bits |= int(i)
		}

		if bits&U_REMOVE == U_REMOVE {
			message.traceMessageReadAdditionlInfo("U_REMOVE", 0)
		}

		if bits&U_MODEL == U_MODEL {
			message.traceAddMessageReadTrace("model")
			err, entity.ModelIndex = message.readByte()
			if err != nil {
				return err
			}
			message.traceMessageReadAdditionlInfo("modelname: ", mvd.Server.Modellist[entity.ModelIndex])
		}
		if bits&U_FRAME == U_FRAME {
			message.traceAddMessageReadTrace("frame")
			err, entity.Frame = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_COLORMAP == U_COLORMAP {
			message.traceAddMessageReadTrace("colormap")
			err, entity.ColorMap = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_SKIN == U_SKIN {
			message.traceAddMessageReadTrace("skin")
			err, b := message.readByte()
			if err != nil {
				return err
			}
			entity.SkinNum = b
		}
		if bits&U_EFFECTS == U_EFFECTS {
			message.traceAddMessageReadTrace("effect")
			err, b := message.readByte()
			if err != nil {
				return err
			}
			entity.Effects = b
		}
		if bits&U_ORIGIN1 == U_ORIGIN1 {
			message.traceAddMessageReadTrace("Origin.X")
			err, entity.Origin.X = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE1 == U_ANGLE1 {
			message.traceAddMessageReadTrace("Angle.X")
			err, entity.Angle.X = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN2 == U_ORIGIN2 {
			message.traceAddMessageReadTrace("Origin.Y")
			err, entity.Origin.Y = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE2 == U_ANGLE2 {
			message.traceAddMessageReadTrace("Angle.Y")
			err, entity.Angle.Y = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN3 == U_ORIGIN3 {
			message.traceAddMessageReadTrace("Origin.Z")
			err, entity.Origin.Z = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE3 == U_ANGLE3 {
			message.traceAddMessageReadTrace("Angle.Z")
			err, entity.Angle.Z = message.readAngle()
			if err != nil {
				return err
			}
		}
		mvd.State.Entities = append(mvd.State.Entities, *entity)
	}
	return nil
}

func (message *Message) Svc_temp_entity(mvd *Mvd) error {
	entity := new(Tempentity)
	message.traceAddMessageReadTrace("flags")
	err, b := message.readByte()
	if err != nil {
		return err
	}

	t := TE_TYPE(b)
	entity.Type = t

	if t == TE_GUNSHOT || t == TE_BLOOD {
		message.traceAddMessageReadTrace("count")
		err, entity.Count = message.readByteAsInt()
		if err != nil {
			return err
		}
	}

	if t == TE_LIGHTNING1 || t == TE_LIGHTNING2 || t == TE_LIGHTNING3 {
		message.traceAddMessageReadTrace("beam-entity")
		err, entity.Entity = message.readShort()
		if err != nil {
			return err
		}
		message.traceAddMessageReadTrace("beam-Start.X")
		err, entity.Start.X = message.readCoord()
		if err != nil {
			return err
		}
		message.traceAddMessageReadTrace("beam-Start.Y")
		err, entity.Start.Y = message.readCoord()
		if err != nil {
			return err
		}
		message.traceAddMessageReadTrace("beam-Start.Z")
		err, entity.Start.Z = message.readCoord()
		if err != nil {
			return err
		}
	}

	message.traceAddMessageReadTrace("Origin.X")
	err, entity.Origin.X = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Y")
	err, entity.Origin.Y = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Z")
	err, entity.Origin.Z = message.readCoord()
	if err != nil {
		return err
	}

	mvd.State.TempEntities = append(mvd.State.TempEntities, *entity)
	return nil
}

func (message *Message) Svc_print(mvd *Mvd) error {
	message.traceAddMessageReadTrace("from")
	err, from := message.readByte()
	if err != nil {
		return err
	}

	message.traceAddMessageReadTrace("message")
	err, s := message.readString()
	if err != nil {
		return err
	}
	mvd.State.Messages = append(mvd.State.Messages, ServerMessage{int(from), s})
	return nil
}

func (message *Message) Svc_serverinfo(mvd *Mvd) error {
	message.traceAddMessageReadTrace("key")
	err, key := message.readString()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("value")
	err, value := message.readString()
	if err != nil {
		return err
	}
	if key == "hostname" {
		mvd.Server.Hostname = value
	}
	mvd.State.Serverinfo = append(mvd.State.Serverinfo, Serverinfo{key, value})
	mvd.Server.Serverinfo[key] = value
	return nil
}

func mvdPrint(s ...interface{}) {
}

func (message *Message) Svc_centerprint(mvd *Mvd) error {
	message.traceAddMessageReadTrace("message")
	err, s := message.readString()
	if err != nil {
		return err
	}
	mvd.State.Centerprint = append(mvd.State.Centerprint, s)
	return nil
}

func (message *Message) Svc_setinfo(mvd *Mvd) error {
	message.traceAddMessageReadTrace("pnum")
	err, pnum := message.readByte() // num
	if err != nil {
		return err
	}

	message.traceAddMessageReadTrace("key")
	err, key := message.readString() // key
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("value")
	err, value := message.readString() // value
	if err != nil {
		return err
	}
	mvd.State.Players[pnum].Setinfo[key] = value
	mvd.emitEventPlayer(&mvd.State.Players[int(pnum)], pnum, PE_USERINFO)
	return nil
}

func (message *Message) Svc_damage(mvd *Mvd) error {
	message.traceAddMessageReadTrace("armor")
	err, _ := message.readByte() // armor
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("blood")
	err, _ = message.readByte() // blood
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.X")
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Y")
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Z")
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_chokecount(mvd *Mvd) error {
	message.traceAddMessageReadTrace("chokecount")
	err, _ := message.readByte()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) parseBaseline(mvd *Mvd) (*Entity, error) {
	var err error
	entity := new(Entity)
	message.traceAddMessageReadTrace("entity-ModelIndex")
	err, entity.ModelIndex = message.readByte()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-ModelFrame")
	err, entity.Frame = message.readByte()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-ColorMap")
	err, entity.ColorMap = message.readByte()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-SkinNum")
	err, entity.SkinNum = message.readByte()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Origin.X")
	err, entity.Origin.X = message.readCoord()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Angle.X")
	err, entity.Angle.X = message.readAngle()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Origin.Y")
	err, entity.Origin.Y = message.readCoord()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Angle.Y")
	err, entity.Angle.Y = message.readAngle()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Origin.Z")
	err, entity.Origin.Z = message.readCoord()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Angle.Z")
	err, entity.Angle.Z = message.readAngle()
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (message *Message) Svc_spawnstatic(mvd *Mvd) error {
	entity, err := message.parseBaseline(mvd)
	if err != nil {
		return err
	}
	mvd.Server.StaticEntities = append(mvd.Server.StaticEntities, *entity)
	return nil
}

func (message *Message) Nq_svc_cutscene(mvd *Mvd) error {
	return message.Svc_smallkick(mvd)
}

func (message *Message) Svc_smallkick(mvd *Mvd) error {
	return nil
}

func (message *Message) Svc_bigkick(mvd *Mvd) error {
	return nil
}

func (message *Message) Svc_muzzleflash(mvd *Mvd) error {
	message.traceAddMessageReadTrace("no_idea")
	err, _ := message.readShort()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_intermission(mvd *Mvd) error {
	message.traceAddMessageReadTrace("Origin.X")
	err, _ := message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Y")
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Z")
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.X")
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.Y")
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.Z")
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_disconnect(mvd *Mvd) error {
	mvd.done = true
	return nil
}

func (message *Message) Svc_setpause(mvd *Mvd) error {
	err, _ := message.readByte()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_foundsecret(mvd *Mvd) error {
	// TODO: maybe add this to the playerinfo
	return nil
}

func (message *Message) Svc_fte_modellistshort(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index_start")
	err, _ := message.readShort() // those are some indexes
	if err != nil {
		return err
	}
	for {
		message.traceAddMessageReadTrace("name")
		err, s := message.readString()
		if err != nil {
			return err
		}
		if len(s) == 0 {
			break
		}
		message.mvd.Server.Modellist = append(message.mvd.Server.Modellist, s)
	}
	message.traceAddMessageReadTrace("offset")
	err, _ = message.readByte() // some more indexes
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_fte_spawnbaseline2(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index")
	err, bits := message.readShort()
	message.traceMessageReadAdditionlInfo("num", bits&511)
	entNum := bits & 511
	if err != nil {
		return err
	}

	bits &= ^511
	if bits&U_MOREBITS == U_MOREBITS {
		message.traceAddMessageReadTrace("morebits")
		message.traceMessageReadAdditionlInfo("morebits happaned?", "?")
		err, i := message.readByte()

		if err != nil {
			return err
		}
		bits |= int(i)
	}

	morebits := 0
	if bits&U_FTE_EVENVENMORE == U_FTE_EVENVENMORE {
		err, i := message.readByte()
		if err != nil {
			return err
		}
		morebits = int(i)
		if morebits&U_FTE_YETMORE == U_FTE_YETMORE {
			err, mi := message.readByte()
			if err != nil {
				return err
			}
			morebits = morebits | int(mi)<<8

		}
	}

	entity := new(Entity)
	entity.Index = entNum

	if bits&U_MODEL == U_MODEL {
		message.traceAddMessageReadTrace("model")
		err, entity.ModelIndex = message.readByte()
		if err != nil {
			return err
		}
	}
	if bits&U_FRAME == U_FRAME {
		message.traceAddMessageReadTrace("frame")
		err, entity.Frame = message.readByte()
		if err != nil {
			return err
		}
	}
	if bits&U_COLORMAP == U_COLORMAP {
		message.traceAddMessageReadTrace("colormap")
		err, entity.ColorMap = message.readByte()
		if err != nil {
			return err
		}
	}
	if bits&U_SKIN == U_SKIN {
		message.traceAddMessageReadTrace("skin")
		err, entity.SkinNum = message.readByte()
		if err != nil {
			return err
		}
	}
	if bits&U_EFFECTS == U_EFFECTS {
		message.traceAddMessageReadTrace("effects")
		err, entity.Effects = message.readByte()
		if err != nil {
			return err
		}
	}
	if bits&U_ORIGIN1 == U_ORIGIN1 {
		message.traceAddMessageReadTrace("Origin.X")
		err, entity.Origin.X = message.readCoord()
		if err != nil {
			return err
		}
	}
	if bits&U_ANGLE1 == U_ANGLE1 {
		message.traceAddMessageReadTrace("Angle.X")
		err, entity.Angle.X = message.readAngle()
		if err != nil {
			return err
		}
	}
	if bits&U_ORIGIN2 == U_ORIGIN2 {
		message.traceAddMessageReadTrace("Origin.Y")
		err, entity.Origin.Y = message.readCoord()
		if err != nil {
			return err
		}
	}
	if bits&U_ANGLE2 == U_ANGLE2 {
		message.traceAddMessageReadTrace("Angle.Y")
		err, entity.Angle.Y = message.readAngle()
		if err != nil {
			return err
		}
	}
	if bits&U_ORIGIN3 == U_ORIGIN3 {
		message.traceAddMessageReadTrace("Origin.Z")
		err, entity.Origin.Z = message.readCoord()
		if err != nil {
			return err
		}
	}
	if bits&U_ANGLE3 == U_ANGLE3 {
		message.traceAddMessageReadTrace("Angle.Z")
		err, entity.Angle.Z = message.readAngle()
		if err != nil {
			return err
		}
	}
	if morebits&U_FTE_TRANS == U_FTE_TRANS {
		err, t := message.readByte()
		if err != nil {
			return err
		}
		entity.Transparency = t
	}
	if morebits&U_FTE_ENTITYDBL == U_FTE_ENTITYDBL {
		entity.Index += 512
	}
	if morebits&U_FTE_ENTITYDBL2 == U_FTE_ENTITYDBL2 {
		entity.Index += 1024
	}
	mvd.Server.Baseline = append(mvd.Server.Baseline, entity)
	return nil
}

// FIXME: this should never happen
func (message *Message) Svc_setview(mvd *Mvd) error {
	err, _ := message.readShort()
	return err
}

func (message *Message) readBytes(count uint) (error, *bytes.Buffer) {
	message.traceStartMessageReadTrace("readBytes", &message.offset, nil, nil)
	if message.offset+count > message.size {
		return errors.New("reading beyong message length"), nil
	}
	b := bytes.NewBuffer(message.data[message.offset : message.offset+count])
	message.offset += count
	message.traceStartMessageReadTrace("readBytes", nil, &message.offset, b)
	return nil, b
}

func (message *Message) readByte() (error, byte) {
	var b byte
	message.traceStartMessageReadTrace("readByte", &message.offset, nil, nil)
	err, barray := message.readBytes(1)
	if err != nil {
		return err, byte(0)
	}
	err = binary.Read(barray, binary.BigEndian, &b)
	if err != nil {
		return err, byte(0)
	}
	message.traceStartMessageReadTrace("readByte", nil, &message.offset, b)
	return nil, b
}

func (message *Message) readByteAsInt() (error, int) {
	message.traceStartMessageReadTrace("readByteAsInt", &message.offset, nil, nil)
	err, b := message.readByte()
	message.traceStartMessageReadTrace("readByteAsInt", nil, &message.offset, int(b))
	return err, int(b)
}

func (message *Message) readLong() (error, int) {
	var i int32
	message.traceStartMessageReadTrace("readLong", &message.offset, nil, nil)
	err, b := message.readBytes(4)
	if err != nil {
		return err, 0
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}
	message.traceStartMessageReadTrace("readLong", nil, &message.offset, i)
	return nil, int(i)
}

func (message *Message) readFloat() (error, float32) {
	var i float32
	message.traceStartMessageReadTrace("readFloat", &message.offset, nil, nil)
	err, b := message.readBytes(4)
	if err != nil {
		return err, 0
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}

	message.traceStartMessageReadTrace("readFloat", nil, &message.offset, float32(i))
	return nil, float32(i)
}

func (message *Message) readString() (error, string) {
	b := make([]byte, 0)
	message.traceStartMessageReadTrace("readString", &message.offset, nil, nil)
	for {
		err, c := message.readByte()
		if err != nil {
			return err, ""
		}
		if c == 255 {
			continue
		}
		if c == 0 {
			break
		}
		b = append(b, c)
	}

	message.traceStartMessageReadTrace("readString", nil, &message.offset, string(b))
	return nil, string(b)
}

func (message *Message) readCoord() (error, float32) {
	message.traceStartMessageReadTrace("readCoord", &message.offset, nil, nil)
	if message.mvd.demo.fte_pext&FTE_PEXT_FLOATCOORDS == FTE_PEXT_FLOATCOORDS {
		err, f := message.readFloat()
		if err != nil {
			return err, 0
		}
		message.traceStartMessageReadTrace("readCoord", nil, &message.offset, f)
		return nil, f
	}
	err, b := message.readShort()
	if err != nil {
		return err, 0
	}

	message.traceStartMessageReadTrace("readCoord", nil, &message.offset, float32(b)*(1.0/8))
	return nil, float32(b) * (1.0 / 8)
}

func (message *Message) readAngle() (error, float32) {
	message.traceStartMessageReadTrace("readAngle", &message.offset, nil, nil)
	if message.mvd.demo.fte_pext&FTE_PEXT_FLOATCOORDS == FTE_PEXT_FLOATCOORDS {

		err, a := message.readAngle16()
		if err != nil {
			return err, 0
		}
		message.traceStartMessageReadTrace("readAngle", nil, &message.offset, a)
		return nil, a
	}
	err, b := message.readByte()
	if err != nil {
		return err, 0
	}
	message.traceStartMessageReadTrace("readAngle", nil, &message.offset, float32(b)*(360.0/256.0))
	return nil, float32(b) * (360.0 / 256.0)
}

func (message *Message) readAngle16() (error, float32) {
	message.traceStartMessageReadTrace("readAngle16", &message.offset, nil, nil)
	err, b := message.readShort()
	if err != nil {
		return err, 0
	}

	message.traceStartMessageReadTrace("readAngle16", nil, &message.offset, float32(b)*(360.0/65536))
	return nil, float32(b) * (360.0 / 65536)
}

func (message *Message) readShort() (error, int) {
	var i int16
	message.traceStartMessageReadTrace("readShort", &message.offset, nil, nil)
	err, b := message.readBytes(2)
	if err != nil {
		return err, 0
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}
	message.traceStartMessageReadTrace("readShort", nil, &message.offset, int(i))
	return nil, int(i)
}

func (message *Message) peekBytes(count, offset uint) (error, *bytes.Buffer) {
	offs := new(uint)
	*offs = message.offset + offset
	message.traceStartMessageReadTrace("peekBytes", &message.offset, nil, nil)
	if message.offset+count > message.size {
		return errors.New("reading beyond message length"), nil
	}
	b := bytes.NewBuffer(message.data[message.offset : message.offset+count])
	offs = new(uint)
	*offs = message.offset + offset + count
	message.traceStartMessageReadTrace("peekBytes", nil, offs, b)
	return nil, b
}

func (message *Message) peekByte(offset uint) (error, byte) {
	var b byte
	offs := new(uint)
	*offs = message.offset + offset
	message.traceStartMessageReadTrace("peekByte", offs, nil, nil)
	err, barray := message.peekBytes(1, offset)
	if err != nil {
		return err, byte(0)
	}
	err = binary.Read(barray, binary.BigEndian, &b)
	if err != nil {
		return err, byte(0)
	}
	offs = new(uint)
	*offs = message.offset + offset + 1
	message.traceStartMessageReadTrace("peekByte", nil, offs, b)
	return nil, b
}
