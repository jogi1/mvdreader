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
	offset      uint
	OffsetStart uint
	OffsetStop  uint
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

func (mvd *Mvd) messageParse(message Message) (bool, error) {
	message.mvd = mvd
	for {
		if mvd.done {
			return true, nil
		}

		mvd.traceStartMessageTrace(&message)
		mvd.traceStartMessageTraceReadTrace("type")

		msgt, err := message.readByte()
		if err != nil {
			return false, err
		}
		msg_type := SVC_TYPE(msgt)
        fmt.Println(msg_type)
		mvd.traceMessageInfo(msg_type)

		if mvd.debug != nil {
			mvd.debug.Println("handling: ", msg_type)
			mvd.debug.Println("expected function: ", strings.Title(msg_type.String()))
		}
		m := reflect.ValueOf(&message).MethodByName(strings.Title(msg_type.String()))

		if m.IsValid() {
			m.Call([]reflect.Value{reflect.ValueOf(mvd)})
		} else {
			return false, fmt.Errorf("error for message type: %#v %#v", msg_type, m)
		}
		if message.offset >= message.size {
			return true, nil
		}

		if mvd.done {
			return true, nil
		}
	}
	if message.offset != message.size {
		return false, errors.New(fmt.Sprint("did not read message fully ", message.offset, message.size))
	}
	return true, nil
}

func (message *Message) Svc_serverdata(mvd *Mvd) error {
	var mrt *TraceRead
	for {
		message.traceAddMessageReadTrace("protocol")
		prot, err := message.readLong()
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
			fte_pext2, err := message.readLong()
			if err != nil {
				return err
			}
			message.mvd.demo.fte_pext2 = FTE_PROTOCOL_EXTENSION(fte_pext2)
			continue
		} else if protocol == protocol_fte {
			message.traceAddMessageReadTrace("protocol_fte")
			fte_pext, err := message.readLong()
			if err != nil {
				return err
			}
			message.mvd.demo.fte_pext = FTE_PROTOCOL_EXTENSION(fte_pext)
			continue
		} else if protocol == protocol_mvd1 {
			message.traceAddMessageReadTrace("protocol_mvd")
			mvd_pext, err := message.readLong()
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
	server_count, err := message.readLong() // server count
	if err != nil {
		return err
	}
	mvd.Server.ServerCount = server_count

	message.traceAddMessageReadTrace("gamedir")
	gamedir, err := message.readString() // gamedir
	if err != nil {
		return err
	}
	mvd.Server.Gamedir = gamedir

	message.traceAddMessageReadTrace("demotime")
	demotime, err := message.readFloat() // demotime
	if err != nil {
		fmt.Println(err)
		return err
	}
	mvd.Server.Demotime = demotime

	message.traceAddMessageReadTrace("map")
	s, err := message.readString()
	if err != nil {
		return err
	}
	mvd.Server.Mapname = s
	for i := 0; i < 10; i++ {

		message.traceAddMessageReadTrace(fmt.Sprintf("movevar_%d", i))
		mv, err := message.readFloat()
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
	_, err := message.readByte()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_stufftext(mvd *Mvd) error {
	message.traceAddMessageReadTrace("stufftext")
	s, err := message.readString()
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
	_, err := message.readByte() // those are some indexes
	if err != nil {
		return err
	}
	for {
		message.traceAddMessageReadTrace("name")
		s, err := message.readString()
		if err != nil {
			return err
		}
		if len(s) == 0 {
			break
		}
		message.mvd.Server.Soundlist = append(message.mvd.Server.Soundlist, s)
	}

	message.traceAddMessageReadTrace("offset")
	_, err = message.readByte() // some more indexes

	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_modellist(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index_start")
	_, err := message.readByte() // those are some indexes
	if err != nil {
		return err
	}
	for {
		message.traceAddMessageReadTrace("name")
		s, err := message.readString()
		if err != nil {
			return err
		}
		if len(s) == 0 {
			break
		}
		message.mvd.Server.Modellist = append(message.mvd.Server.Modellist, s)
	}
	message.traceAddMessageReadTrace("offset")
	_, err = message.readByte() // some more indexes
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_spawnbaseline(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index")
	index, err := message.readShort()
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
	player, err := message.readByte()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("frags")
	frags, err := message.readShort()
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
	pnum, err := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]

	message.traceAddMessageReadTrace("flags")
	sflags, err := message.readShort()
	if err != nil {
		return err
	}
	flags := DF_TYPE(sflags)

	message.traceAddMessageReadTrace("frame")
	frame, err := message.readByte()
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
					coord, err := message.readCoord()
					if err != nil {
						return err
					}
					p.Origin.X = coord
				}
			case 1:
				{
					message.traceAddMessageReadTrace("Origin.Y")
					coord, err := message.readCoord()
					if err != nil {
						return err
					}
					p.Origin.Y = coord
				}
			case 2:
				{
					message.traceAddMessageReadTrace("Origin.Z")
					coord, err := message.readCoord()
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
					angle, err := message.readAngle16()
					if err != nil {
						return err
					}
					p.Angle.X = angle
				}
			case 1:
				{
					message.traceAddMessageReadTrace("Angle.Y")
					angle, err := message.readAngle16()
					if err != nil {
						return err
					}
					p.Angle.Y = angle
				}
			case 2:
				{
					message.traceAddMessageReadTrace("Angle.Z")
					angle, err := message.readAngle16()
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
		mindex, err := message.readByte()
		if err != nil {
			return err
		}
		p.ModelIndex = mindex // modelindex
	}

	if flags&DF_SKINNUM == DF_SKINNUM {
		pe_type |= PE_ANIMATION
		message.traceAddMessageReadTrace("skinnum")
		skinnum, err := message.readByte()
		if err != nil {
			return err
		}
		p.SkinNum = skinnum // skinnum
	}

	if flags&DF_EFFECTS == DF_EFFECTS {
		pe_type |= PE_ANIMATION
		message.traceAddMessageReadTrace("effects")
		effects, err := message.readByte()
		if err != nil {
			return err
		}
		p.Effects = effects // effects
	}

	if flags&DF_WEAPONFRAME == DF_WEAPONFRAME {
		pe_type |= PE_ANIMATION

		message.traceAddMessageReadTrace("weaponframe")
		weaponframe, err := message.readByte()
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
	pnum, err := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]

	message.traceAddMessageReadTrace("ping")
	ping, err := message.readShort()
	if err != nil {
		return err
	}
	p.Ping = ping // ping
	mvd.emitEventPlayer(p, pnum, PE_NETWORKINFO)
	return nil
}

func (message *Message) Svc_updatepl(mvd *Mvd) error {
	message.traceAddMessageReadTrace("pnum")
	pnum, err := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]

	message.traceAddMessageReadTrace("pl")
	pl, err := message.readByte()
	if err != nil {
		return err
	}
	p.Pl = pl // pl
	mvd.emitEventPlayer(p, pnum, PE_NETWORKINFO)
	return nil
}

func (message *Message) Svc_updateentertime(mvd *Mvd) error {
	message.traceAddMessageReadTrace("pnum")

	pnum, err := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]

	message.traceAddMessageReadTrace("entertime")
	entertime, err := message.readFloat()
	if err != nil {
		return err
	}
	p.Entertime = entertime // entertime
	mvd.emitEventPlayer(p, pnum, PE_NETWORKINFO)
	return nil
}

func (message *Message) Svc_updateuserinfo(mvd *Mvd) error {
	message.traceAddMessageReadTrace("pnum")
	pnum, err := message.readByte()
	if err != nil {
		return err
	}

	message.traceAddMessageReadTrace("uid")
	uid, err := message.readLong()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]
	p.Userid = uid

	message.traceAddMessageReadTrace("userinfo")
	ui, err := message.readString()
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
	sc, err := message.readShort()
	if err != nil {
		return err
	}
	channel := SND_TYPE(sc) // channel
	s.Channel = channel
	if channel&SND_VOLUME == SND_VOLUME {
		message.traceAddMessageReadTrace("volume")
		volume, err := message.readByte()
		if err != nil {
			return err
		}
		s.Volume = volume
	}

	if channel&SND_ATTENUATION == SND_ATTENUATION {
		message.traceAddMessageReadTrace("attenuation")
		attenuation, err := message.readByte()
		if err != nil {
			return err
		}
		s.Attenuation = attenuation
	}
	ent := (s.Channel >> 3) & 1023
	message.traceMessageReadAdditionlInfo("ent", ent)
	message.traceMessageReadAdditionlInfo("actual channel", s.Channel&7)
	message.traceAddMessageReadTrace("index")
	index, err := message.readByte()
	if err != nil {
		return err
	}
	s.Index = index // sound_num

	message.traceAddMessageReadTrace("Origin.X")
	x, err := message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Y")
	y, err := message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Z")
	z, err := message.readCoord()
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
	x, err := message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Y")
	y, err := message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Z")
	z, err := message.readCoord()
	if err != nil {
		return err
	}
	s.Origin.Set(x, y, z)

	message.traceAddMessageReadTrace("index")
	index, err := message.readByte()
	if err != nil {
		return err
	}
	s.Index = index // sound_num

	message.traceAddMessageReadTrace("volume")
	volume, err := message.readByte()
	if err != nil {
		return err
	}
	s.Volume = volume // sound volume
	message.traceAddMessageReadTrace("attenuation")
	attenuation, err := message.readByte()
	if err != nil {
		return err
	}
	s.Attenuation = attenuation // sound attenuation
	mvd.State.SoundsStatic = append(mvd.State.SoundsStatic, s)
	return nil
}

func (message *Message) Svc_setangle(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index")
	_, err := message.readByte() // something weird?
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.X")
	_, err = message.readAngle() // x
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.Y")
	_, err = message.readAngle()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.Z")
	_, err = message.readAngle()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_lightstyle(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index")
	_, err := message.readByte() // lightstyle num
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("style")
	_, err = message.readString()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_updatestatlong(mvd *Mvd) error {
	message.traceAddMessageReadTrace("stat")
	b, err := message.readByte()
	if err != nil {
		return err
	}
	stat := STAT_TYPE(b)
	message.traceAddMessageReadTrace("value")
	value, err := message.readLong()
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
	}
	mvd.emitEventPlayer(p, byte(mvd.demo.last_to), PE_STATS)
	return nil
}

func (message *Message) Svc_updatestat(mvd *Mvd) error {
	message.traceAddMessageReadTrace("stat")
	b, err := message.readByte()
	if err != nil {
		return err
	}
	stat := STAT_TYPE(b)

	message.traceAddMessageReadTrace("value")
	value, err := message.readByte()
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
	_, err := message.readByte()
	if err != nil {
		return err
	}

	for {
		message.traceAddMessageReadTrace("flags")
		w, err := message.readShort()
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
			i, err := message.readByte()
			if err != nil {
				return err
			}
			bits |= int(i)
		}

		morebits := 0
		if bits&U_FTE_EVENVENMORE == U_FTE_EVENVENMORE {
			i, err := message.readByte()
			if err != nil {
				return err
			}
			morebits = int(i)
			if morebits&U_FTE_YETMORE == U_FTE_YETMORE {
				mi, err := message.readByte()
				if err != nil {
					return err
				}
				morebits = morebits | int(mi)<<8

			}
		}

		if bits&U_MOREBITS == U_MOREBITS {
			if mvd.demo.fte_pext&FTE_PEXT_ENTITYDBL == FTE_PEXT_ENTITYDBL {
				evenMore, err := message.peekByte(0)
				if err != nil {
					return err
				}
				if evenMore&U_FTE_EVENVENMORE == U_FTE_EVENVENMORE {
					evenMore, err = message.peekByte(1)
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
				ftext, err := message.readByte()
				if err != nil {
					return err
				}
				if ftext&U_FTE_EVENVENMORE == U_FTE_EVENVENMORE {
					_, err := message.readByte()
					if err != nil {
						return err
					}
				}
			}
		}
		if bits&U_MODEL == U_MODEL {
			message.traceAddMessageReadTrace("model")
			entity.ModelIndex, err = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_FRAME == U_FRAME {
			message.traceAddMessageReadTrace("frame")
			entity.Frame, err = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_COLORMAP == U_COLORMAP {
			message.traceAddMessageReadTrace("colormap")
			entity.ColorMap, err = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_SKIN == U_SKIN {
			message.traceAddMessageReadTrace("skin")
			entity.SkinNum, err = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_EFFECTS == U_EFFECTS {
			message.traceAddMessageReadTrace("effects")
			entity.Effects, err = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN1 == U_ORIGIN1 {
			message.traceAddMessageReadTrace("Origin.X")
			entity.Origin.X, err = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE1 == U_ANGLE1 {
			message.traceAddMessageReadTrace("Angle.X")
			entity.Angle.X, err = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN2 == U_ORIGIN2 {
			message.traceAddMessageReadTrace("Origin.Y")
			entity.Origin.Y, err = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE2 == U_ANGLE2 {
			message.traceAddMessageReadTrace("Angle.Y")
			entity.Angle.Y, err = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN3 == U_ORIGIN3 {
			message.traceAddMessageReadTrace("Origin.Z")
			entity.Origin.Z, err = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE3 == U_ANGLE3 {
			message.traceAddMessageReadTrace("Angle.Z")
			entity.Angle.Z, err = message.readAngle()
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
			b, err := message.readByte()
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
		w, err := message.readShort()
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
			i, err := message.readByte()
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
			entity.ModelIndex, err = message.readByte()
			if err != nil {
				return err
			}
			message.traceMessageReadAdditionlInfo("modelname: ", mvd.Server.Modellist[entity.ModelIndex])
		}
		if bits&U_FRAME == U_FRAME {
			message.traceAddMessageReadTrace("frame")
			entity.Frame, err = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_COLORMAP == U_COLORMAP {
			message.traceAddMessageReadTrace("colormap")
			entity.ColorMap, err = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_SKIN == U_SKIN {
			message.traceAddMessageReadTrace("skin")
			b, err := message.readByte()
			if err != nil {
				return err
			}
			entity.SkinNum = b
		}
		if bits&U_EFFECTS == U_EFFECTS {
			message.traceAddMessageReadTrace("effect")
			b, err := message.readByte()
			if err != nil {
				return err
			}
			entity.Effects = b
		}
		if bits&U_ORIGIN1 == U_ORIGIN1 {
			message.traceAddMessageReadTrace("Origin.X")
			entity.Origin.X, err = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE1 == U_ANGLE1 {
			message.traceAddMessageReadTrace("Angle.X")
			entity.Angle.X, err = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN2 == U_ORIGIN2 {
			message.traceAddMessageReadTrace("Origin.Y")
			entity.Origin.Y, err = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE2 == U_ANGLE2 {
			message.traceAddMessageReadTrace("Angle.Y")
			entity.Angle.Y, err = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN3 == U_ORIGIN3 {
			message.traceAddMessageReadTrace("Origin.Z")
			entity.Origin.Z, err = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE3 == U_ANGLE3 {
			message.traceAddMessageReadTrace("Angle.Z")
			entity.Angle.Z, err = message.readAngle()
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
	b, err := message.readByte()
	if err != nil {
		return err
	}

	t := TE_TYPE(b)
	entity.Type = t

	if t == TE_GUNSHOT || t == TE_BLOOD {
		message.traceAddMessageReadTrace("count")
		entity.Count, err = message.readByteAsInt()
		if err != nil {
			return err
		}
	}

	if t == TE_LIGHTNING1 || t == TE_LIGHTNING2 || t == TE_LIGHTNING3 {
		message.traceAddMessageReadTrace("beam-entity")
		entity.Entity, err = message.readShort()
		if err != nil {
			return err
		}
		message.traceAddMessageReadTrace("beam-Start.X")
		entity.Start.X, err = message.readCoord()
		if err != nil {
			return err
		}
		message.traceAddMessageReadTrace("beam-Start.Y")
		entity.Start.Y, err = message.readCoord()
		if err != nil {
			return err
		}
		message.traceAddMessageReadTrace("beam-Start.Z")
		entity.Start.Z, err = message.readCoord()
		if err != nil {
			return err
		}
	}

	message.traceAddMessageReadTrace("Origin.X")
	entity.Origin.X, err = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Y")
	entity.Origin.Y, err = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Z")
	entity.Origin.Z, err = message.readCoord()
	if err != nil {
		return err
	}

	mvd.State.TempEntities = append(mvd.State.TempEntities, *entity)
	return nil
}

func (message *Message) Svc_print(mvd *Mvd) error {
	message.traceAddMessageReadTrace("from")
	from, err := message.readByte()
	if err != nil {
		return err
	}

	message.traceAddMessageReadTrace("message")
	s, err := message.readString()
	if err != nil {
		return err
	}
	mvd.State.Messages = append(mvd.State.Messages, ServerMessage{int(from), s})
	return nil
}

func (message *Message) Svc_serverinfo(mvd *Mvd) error {
	message.traceAddMessageReadTrace("key")
	key, err := message.readString()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("value")
	value, err := message.readString()
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
	s, err := message.readString()
	if err != nil {
		return err
	}
	mvd.State.Centerprint = append(mvd.State.Centerprint, s)
	return nil
}

func (message *Message) Svc_setinfo(mvd *Mvd) error {
	message.traceAddMessageReadTrace("pnum")
	pnum, err := message.readByte() // num
	if err != nil {
		return err
	}

	message.traceAddMessageReadTrace("key")
	key, err := message.readString() // key
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("value")
	value, err := message.readString() // value
	if err != nil {
		return err
	}
	mvd.State.Players[pnum].Setinfo[key] = value
	mvd.emitEventPlayer(&mvd.State.Players[int(pnum)], pnum, PE_USERINFO)
	return nil
}

func (message *Message) Svc_damage(mvd *Mvd) error {
	message.traceAddMessageReadTrace("armor")
	_, err := message.readByte() // armor
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("blood")
	_, err = message.readByte() // blood
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.X")
	_, err = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Y")
	_, err = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Z")
	_, err = message.readCoord()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_chokecount(mvd *Mvd) error {
	message.traceAddMessageReadTrace("chokecount")
	_, err := message.readByte()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) parseBaseline(mvd *Mvd) (*Entity, error) {
	var err error
	entity := new(Entity)
	message.traceAddMessageReadTrace("entity-ModelIndex")
	entity.ModelIndex, err = message.readByte()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-ModelFrame")
	entity.Frame, err = message.readByte()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-ColorMap")
	entity.ColorMap, err = message.readByte()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-SkinNum")
	entity.SkinNum, err = message.readByte()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Origin.X")
	entity.Origin.X, err = message.readCoord()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Angle.X")
	entity.Angle.X, err = message.readAngle()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Origin.Y")
	entity.Origin.Y, err = message.readCoord()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Angle.Y")
	entity.Angle.Y, err = message.readAngle()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Origin.Z")
	entity.Origin.Z, err = message.readCoord()
	if err != nil {
		return nil, err
	}
	message.traceAddMessageReadTrace("entity-Angle.Z")
	entity.Angle.Z, err = message.readAngle()
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
	_, err := message.readShort()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_intermission(mvd *Mvd) error {
	message.traceAddMessageReadTrace("Origin.X")
	_, err := message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Y")
	_, err = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Origin.Z")
	_, err = message.readCoord()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.X")
	_, err = message.readAngle()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.Y")
	_, err = message.readAngle()
	if err != nil {
		return err
	}
	message.traceAddMessageReadTrace("Angle.Z")
	_, err = message.readAngle()
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
	_, err := message.readByte()
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
	_, err := message.readShort() // those are some indexes
	if err != nil {
		return err
	}
	for {
		message.traceAddMessageReadTrace("name")
		s, err := message.readString()
		if err != nil {
			return err
		}
		if len(s) == 0 {
			break
		}
		message.mvd.Server.Modellist = append(message.mvd.Server.Modellist, s)
	}
	message.traceAddMessageReadTrace("offset")
	_, err = message.readByte() // some more indexes
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_fte_spawnbaseline2(mvd *Mvd) error {
	message.traceAddMessageReadTrace("index")
	bits, err := message.readShort()
	message.traceMessageReadAdditionlInfo("num", bits&511)
	entNum := bits & 511
	if err != nil {
		return err
	}

	bits &= ^511
	if bits&U_MOREBITS == U_MOREBITS {
		message.traceAddMessageReadTrace("morebits")
		message.traceMessageReadAdditionlInfo("morebits happaned?", "?")
		i, err := message.readByte()
		if err != nil {
			return err
		}
		bits |= int(i)
	}

	morebits := 0
	if bits&U_FTE_EVENVENMORE == U_FTE_EVENVENMORE {
		i, err := message.readByte()
		if err != nil {
			return err
		}
		morebits = int(i)
		if morebits&U_FTE_YETMORE == U_FTE_YETMORE {
			mi, err := message.readByte()
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
		entity.ModelIndex, err = message.readByte()
		if err != nil {
			return err
		}
	}
	if bits&U_FRAME == U_FRAME {
		message.traceAddMessageReadTrace("frame")
		entity.Frame, err = message.readByte()
		if err != nil {
			return err
		}
	}
	if bits&U_COLORMAP == U_COLORMAP {
		message.traceAddMessageReadTrace("colormap")
		entity.ColorMap, err = message.readByte()
		if err != nil {
			return err
		}
	}
	if bits&U_SKIN == U_SKIN {
		message.traceAddMessageReadTrace("skin")
		entity.SkinNum, err = message.readByte()
		if err != nil {
			return err
		}
	}
	if bits&U_EFFECTS == U_EFFECTS {
		message.traceAddMessageReadTrace("effects")
		entity.Effects, err = message.readByte()
		if err != nil {
			return err
		}
	}
	if bits&U_ORIGIN1 == U_ORIGIN1 {
		message.traceAddMessageReadTrace("Origin.X")
		entity.Origin.X, err = message.readCoord()
		if err != nil {
			return err
		}
	}
	if bits&U_ANGLE1 == U_ANGLE1 {
		message.traceAddMessageReadTrace("Angle.X")
		entity.Angle.X, err = message.readAngle()
		if err != nil {
			return err
		}
	}
	if bits&U_ORIGIN2 == U_ORIGIN2 {
		message.traceAddMessageReadTrace("Origin.Y")
		entity.Origin.Y, err = message.readCoord()
		if err != nil {
			return err
		}
	}
	if bits&U_ANGLE2 == U_ANGLE2 {
		message.traceAddMessageReadTrace("Angle.Y")
		entity.Angle.Y, err = message.readAngle()
		if err != nil {
			return err
		}
	}
	if bits&U_ORIGIN3 == U_ORIGIN3 {
		message.traceAddMessageReadTrace("Origin.Z")
		entity.Origin.Z, err = message.readCoord()
		if err != nil {
			return err
		}
	}
	if bits&U_ANGLE3 == U_ANGLE3 {
		message.traceAddMessageReadTrace("Angle.Z")
		entity.Angle.Z, err = message.readAngle()
		if err != nil {
			return err
		}
	}
	if morebits&U_FTE_TRANS == U_FTE_TRANS {
		t, err := message.readByte()
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
	_, err := message.readShort()
	return err
}

func (message *Message) readBytes(count uint) (*bytes.Buffer, error) {
	message.traceStartMessageReadTrace("readBytes", &message.offset, nil, nil)
	if message.offset+count > message.size {
		return nil, errors.New("reading beyond message length")
	}
	b := bytes.NewBuffer(message.mvd.file[message.OffsetStart+message.offset : message.OffsetStart+message.offset+count])
	message.offset += count
	message.traceStartMessageReadTrace("readBytes", nil, &message.offset, b)
	return b, nil
}

func (message *Message) readByte() (byte, error) {
	var b byte
	message.traceStartMessageReadTrace("readByte", &message.offset, nil, nil)
	barray, err := message.readBytes(1)
	if err != nil {
		return byte(0), err
	}
	err = binary.Read(barray, binary.BigEndian, &b)
	if err != nil {
		return byte(0), err
	}
	message.traceStartMessageReadTrace("readByte", nil, &message.offset, b)
	return b, nil
}

func (message *Message) readByteAsInt() (int, error) {
	message.traceStartMessageReadTrace("readByteAsInt", &message.offset, nil, nil)
	b, err := message.readByte()
	message.traceStartMessageReadTrace("readByteAsInt", nil, &message.offset, int(b))
	return int(b), err
}

func (message *Message) readLong() (int, error) {
	var i int32
	message.traceStartMessageReadTrace("readLong", &message.offset, nil, nil)
	b, err := message.readBytes(4)
	if err != nil {
		return 0, err
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return 0, err
	}
	message.traceStartMessageReadTrace("readLong", nil, &message.offset, i)
	return int(i), nil
}

func (message *Message) readFloat() (float32, error) {
	var i float32
	message.traceStartMessageReadTrace("readFloat", &message.offset, nil, nil)
	b, err := message.readBytes(4)
	if err != nil {
		return 0, err
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return 0, err
	}

	message.traceStartMessageReadTrace("readFloat", nil, &message.offset, float32(i))
	return float32(i), nil
}

func (message *Message) readString() (string, error) {
	b := make([]byte, 0)
	message.traceStartMessageReadTrace("readString", &message.offset, nil, nil)
	for {
		c, err := message.readByte()
		if err != nil {
			return "", err
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
	return string(b), nil
}

func (message *Message) readCoord() (float32, error) {
	message.traceStartMessageReadTrace("readCoord", &message.offset, nil, nil)
	if message.mvd.demo.fte_pext&FTE_PEXT_FLOATCOORDS == FTE_PEXT_FLOATCOORDS {
		f, err := message.readFloat()
		if err != nil {
			return 0, err
		}
		message.traceStartMessageReadTrace("readCoord", nil, &message.offset, f)
		return f, nil
	}
	b, err := message.readShort()
	if err != nil {
		return 0, err
	}

	message.traceStartMessageReadTrace("readCoord", nil, &message.offset, float32(b)*(1.0/8))
	return float32(b) * (1.0 / 8), nil
}

func (message *Message) readAngle() (float32, error) {
	message.traceStartMessageReadTrace("readAngle", &message.offset, nil, nil)
	if message.mvd.demo.fte_pext&FTE_PEXT_FLOATCOORDS == FTE_PEXT_FLOATCOORDS {

		a, err := message.readAngle16()
		if err != nil {
			return 0, err
		}
		message.traceStartMessageReadTrace("readAngle", nil, &message.offset, a)
		return a, nil
	}
	b, err := message.readByte()
	if err != nil {
		return 0, err
	}
	message.traceStartMessageReadTrace("readAngle", nil, &message.offset, float32(b)*(360.0/256.0))
	return float32(b) * (360.0 / 256.0), nil
}

func (message *Message) readAngle16() (float32, error) {
	message.traceStartMessageReadTrace("readAngle16", &message.offset, nil, nil)
	b, err := message.readShort()
	if err != nil {
		return 0, err
	}

	message.traceStartMessageReadTrace("readAngle16", nil, &message.offset, float32(b)*(360.0/65536))
	return float32(b) * (360.0 / 65536), nil
}

func (message *Message) readShort() (int, error) {
	var i int16
	message.traceStartMessageReadTrace("readShort", &message.offset, nil, nil)
	b, err := message.readBytes(2)
	if err != nil {
		return 0, err
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return 0, err
	}
	message.traceStartMessageReadTrace("readShort", nil, &message.offset, int(i))
	return int(i), nil
}

func (message *Message) peekBytes(count, offset uint) (*bytes.Buffer, error) {
	offs := new(uint)
	*offs = message.offset + offset
	message.traceStartMessageReadTrace("peekBytes", &message.offset, nil, nil)
	if message.offset+count > message.size {
		return nil, errors.New("reading beyond message length")
	}

	b := bytes.NewBuffer(message.mvd.file[message.OffsetStart+message.offset : message.OffsetStart+message.offset+count])
	offs = new(uint)
	*offs = message.offset + offset + count
	message.traceStartMessageReadTrace("peekBytes", nil, offs, b)
	return b, nil
}

func (message *Message) peekByte(offset uint) (byte, error) {
	var b byte
	offs := new(uint)
	*offs = message.offset + offset
	message.traceStartMessageReadTrace("peekByte", offs, nil, nil)
	barray, err := message.peekBytes(1, offset)
	if err != nil {
		return byte(0), err
	}
	err = binary.Read(barray, binary.BigEndian, &b)
	if err != nil {
		return byte(0), err
	}
	offs = new(uint)
	*offs = message.offset + offset + 1
	message.traceStartMessageReadTrace("peekByte", nil, offs, b)
	return b, nil
}
