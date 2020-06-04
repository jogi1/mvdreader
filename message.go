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
	size   uint
	data   []byte
	offset uint
	mvd    *Mvd
}

func (mvd *Mvd) emitEventPlayer(player *Player, pnum byte, pe_type PE_TYPE) {

	player.EventInfo.Pnum = pnum
	player.EventInfo.Events |= pe_type
}

func (mvd *Mvd) emitEventSound(sound *Sound) {
	sound.Frame = mvd.Frame
	mvd.State.SoundsActive = append(mvd.State.SoundsActive, *sound)
}

func (mvd *Mvd) messageParse(message Message) error {
	message.mvd = mvd
	for {
		if mvd.done == true {
			return nil
		}

		err, msgt := message.readByte()
		if err != nil {
			return err
		}
		msg_type := SVC_TYPE(msgt)

		if mvd.debug != nil {
			mvd.debug.Println("handling: ", msg_type)
		}
		m := reflect.ValueOf(&message).MethodByName(strings.Title(fmt.Sprintf("%s", msg_type)))

		if m.IsValid() == true {
			m.Call([]reflect.Value{reflect.ValueOf(mvd)})
		} else {
			return errors.New(fmt.Sprint("error for message type: %#v %#v", msg_type, m))
		}
		if message.offset >= message.size {
			return nil
		}
		if mvd.done {
			return nil
		}
	}
	if message.offset != message.size {
		return errors.New(fmt.Sprint("did not read message fully ", message.offset, message.size))
	}
	return nil
}

func (message *Message) Svc_serverdata(mvd *Mvd) error {
	for {

		err, prot := message.readLong()
		if err != nil {
			return err
		}
		message.mvd.demo.protocol = PROTOCOL_VERSION(prot)
		protocol := message.mvd.demo.protocol

		if protocol == protocol_fte2 {
			err, fte_pext2 := message.readLong()
			if err != nil {
				return err
			}
			message.mvd.demo.fte_pext2 = FTE_PROTOCOL_EXTENSION(fte_pext2)
			continue
		}

		if protocol == protocol_fte {

			err, fte_pext := message.readLong()
			if err != nil {
				return err
			}
			message.mvd.demo.fte_pext = FTE_PROTOCOL_EXTENSION(fte_pext)
			continue
		}

		if protocol == protocol_mvd1 {

			err, mvd_pext := message.readLong()
			if err != nil {
				return err
			}
			message.mvd.demo.mvd_pext = MVD_PROTOCOL_EXTENSION(mvd_pext)
			continue
		}
		if protocol == protocol_standard {
			break
		}
	}

	err, server_count := message.readLong() // server count
	if err != nil {
		return err
	}
	mvd.Server.ServerCount = server_count

	err, gamedir := message.readString() // gamedir
	if err != nil {
		return err
	}
	mvd.Server.Gamedir = gamedir

	err, demotime := message.readFloat() // demotime
	if err != nil {
		return err
	}
	mvd.Server.Demotime = demotime

	err, s := message.readString()
	if err != nil {
		return err
	}
	mvd.Server.Mapname = s
	for i := 0; i < 10; i++ {
		//fmt.Printf("movevar(%v): %v\n", i, message.ReadFloat())
		err, mv := message.readFloat()
		if err != nil {
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
	err, _ := message.readByte()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_stufftext(mvd *Mvd) error {
	err, _ := message.readString()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_soundlist(mvd *Mvd) error {
	err, _ := message.readByte() // those are some indexes
	if err != nil {
		return err
	}
	for {
		err, s := message.readString()
		if err != nil {
			return err
		}
		message.mvd.Server.Soundlist = append(message.mvd.Server.Soundlist, s)
		if len(s) == 0 {
			break
		}
	}
	err, _ = message.readByte() // some more indexes
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_modellist(mvd *Mvd) error {
	err, _ := message.readByte() // those are some indexes
	if err != nil {
		return err
	}
	for {
		err, s := message.readString()
		if err != nil {
			return err
		}
		message.mvd.Server.Modellist = append(message.mvd.Server.Modellist, s)
		if len(s) == 0 {
			break
		}
	}

	err, _ = message.readByte() // some more indexes
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_spawnbaseline(mvd *Mvd) error {
	err, _ := message.readShort() // entity
	if err != nil {
		return err
	}
	err, _ = message.readByte() // modellindex
	if err != nil {
		return err
	}

	err, _ = message.readByte() // frame
	if err != nil {
		return err
	}

	err, _ = message.readByte() // colormap
	if err != nil {
		return err
	}

	err, _ = message.readByte() // skinnum
	if err != nil {
		return err
	}

	for i := 0; i < 3; i++ {
		err, _ = message.readCoord() // coord
		if err != nil {
			return err
		}
		err, _ = message.readAngle() // coord
		if err != nil {
			return err
		}
	}
	return nil
}

func (message *Message) Svc_updatefrags(mvd *Mvd) error {
	err, player := message.readByte()
	if err != nil {
		return err
	}
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
	err, pnum := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]

	err, sflags := message.readShort()
	if err != nil {
		return err
	}
	flags := DF_TYPE(sflags)
	err, frame := message.readByte()
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

					err, coord := message.readCoord()
					if err != nil {
						return err
					}
					p.Origin.X = coord
				}
			case 1:
				{
					err, coord := message.readCoord()
					if err != nil {
						return err
					}
					p.Origin.Y = coord
				}
			case 2:
				{
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
					err, angle := message.readAngle16()
					if err != nil {
						return err
					}
					p.Angle.X = angle
				}
			case 1:
				{
					err, angle := message.readAngle16()
					if err != nil {
						return err
					}
					p.Angle.Y = angle
				}
			case 2:
				{
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
		err, mindex := message.readByte()
		if err != nil {
			return err
		}
		p.ModelIndex = mindex // modelindex
	}

	if flags&DF_SKINNUM == DF_SKINNUM {
		pe_type |= PE_ANIMATION
		err, skinnum := message.readByte()
		if err != nil {
			return err
		}
		p.SkinNum = skinnum // skinnum
	}

	if flags&DF_EFFECTS == DF_EFFECTS {
		pe_type |= PE_ANIMATION
		err, effects := message.readByte()
		if err != nil {
			return err
		}
		p.Effects = effects // effects
	}

	if flags&DF_WEAPONFRAME == DF_WEAPONFRAME {
		pe_type |= PE_ANIMATION
		err, weaponframe := message.readByte()
		if err != nil {
			return err
		}
		p.WeaponFrame = weaponframe // weaponframe
	}

	mvd.emitEventPlayer(p, pnum, pe_type)
	return nil
}

func (message *Message) Svc_updateping(mvd *Mvd) error {
	err, pnum := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]
	err, ping := message.readShort()
	if err != nil {
		return err
	}
	p.Ping = ping // ping
	mvd.emitEventPlayer(p, pnum, PE_NETWORKINFO)
	return nil
}

func (message *Message) Svc_updatepl(mvd *Mvd) error {
	err, pnum := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]
	err, pl := message.readByte()
	if err != nil {
		return err
	}
	p.Pl = pl // pl
	mvd.emitEventPlayer(p, pnum, PE_NETWORKINFO)
	return nil
}

func (message *Message) Svc_updateentertime(mvd *Mvd) error {
	err, pnum := message.readByte()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]

	err, entertime := message.readFloat()
	if err != nil {
		return err
	}
	p.Entertime = entertime // entertime
	mvd.emitEventPlayer(p, pnum, PE_NETWORKINFO)
	return nil
}

func (message *Message) Svc_updateuserinfo(mvd *Mvd) error {

	err, pnum := message.readByte()
	if err != nil {
		return err
	}

	err, uid := message.readLong()
	if err != nil {
		return err
	}
	p := &mvd.State.Players[pnum]
	p.Userid = uid
	err, ui := message.readString()
	if err != nil {
		return err
	}
	if len(ui) < 2 {
		return nil
	}
	ui = ui[1:]
	splits := strings.Split(ui, "\\")
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
	}
	mvd.emitEventPlayer(p, pnum, PE_USERINFO)
	return nil
}

func (message *Message) Svc_sound(mvd *Mvd) error {
	var s Sound
	err, sc := message.readShort()
	if err != nil {
		return err
	}
	channel := SND_TYPE(sc) // channel
	s.Channel = channel
	if channel&SND_VOLUME == SND_VOLUME {
		mvdPrint("has volume")
		err, volume := message.readByte()
		if err != nil {
			return err
		}
		s.Volume = volume
	}

	if channel&SND_ATTENUATION == SND_ATTENUATION {
		mvdPrint("has attenuation")
		err, attenuation := message.readByte()
		if err != nil {
			return err
		}
		s.Attenuation = attenuation
	}

	err, index := message.readByte()
	if err != nil {
		return err
	}
	s.Index = index // sound_num
	err, x := message.readCoord()
	if err != nil {
		return err
	}
	err, y := message.readCoord()
	if err != nil {
		return err
	}
	err, z := message.readCoord()
	if err != nil {
		return err
	}
	s.Origin.Set(x, y, z)
	mvd.State.SoundsActive = append(mvd.State.SoundsActive, s)
	mvd.emitEventSound(&s)
	return nil
}

func (message *Message) Svc_spawnstaticsound(mvd *Mvd) error {
	var s Sound
	err, x := message.readCoord()
	if err != nil {
		return err
	}
	err, y := message.readCoord()
	if err != nil {
		return err
	}
	err, z := message.readCoord()
	if err != nil {
		return err
	}
	s.Origin.Set(x, y, z)

	err, index := message.readByte()
	if err != nil {
		return err
	}
	s.Index = index // sound_num

	err, volume := message.readByte()
	if err != nil {
		return err
	}
	s.Volume = volume // sound volume
	err, attenuation := message.readByte()
	if err != nil {
		return err
	}
	s.Attenuation = attenuation // sound attenuation
	mvd.State.SoundsStatic = append(mvd.State.SoundsStatic, s)
	return nil
}

func (message *Message) Svc_setangle(mvd *Mvd) error {
	err, _ := message.readByte() // something weird?
	if err != nil {
		return err
	}
	err, _ = message.readAngle() // x
	if err != nil {
		return err
	}
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_lightstyle(mvd *Mvd) error {
	err, _ := message.readByte() // lightstyle num
	if err != nil {
		return err
	}
	err, _ = message.readString()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_updatestatlong(mvd *Mvd) error {
	err, b := message.readByte()
	if err != nil {
		return err
	}
	stat := STAT_TYPE(b)
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
	err, b := message.readByte()
	if err != nil {
		return err
	}
	stat := STAT_TYPE(b)
	err, value := message.readByte()
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
	} else {
		return errors.New(fmt.Sprintf("unknown STAT_ type: %s\n", stat))
	}
	mvd.emitEventPlayer(p, byte(mvd.demo.last_to), PE_STATS)
	return nil
}

func (message *Message) Svc_deltapacketentities(mvd *Mvd) error {
	err, from := message.readByte()
	if err != nil {
		return err
	}
	mvdPrint(from)
	for {
		err, w := message.readShort()
		if err != nil {
			return err
		}
		if w == 0 {
			break
		}

		w &= ^511
		bits := w

		if bits&U_MOREBITS == U_MOREBITS {
			err, i := message.readByte()
			if err != nil {
				return err
			}
			bits |= int(i)
		}

		if bits&U_MODEL == U_MODEL {
			err, _ = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_FRAME == U_FRAME {
			err, _ = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_COLORMAP == U_COLORMAP {
			err, _ = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_SKIN == U_SKIN {
			err, _ = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_EFFECTS == U_EFFECTS {
			err, _ = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN1 == U_ORIGIN1 {
			err, _ = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE1 == U_ANGLE1 {
			err, _ = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN2 == U_ORIGIN2 {
			err, _ = message.readCoord()
		}
		if bits&U_ANGLE2 == U_ANGLE2 {
			err, _ = message.readAngle()
		}
		if bits&U_ORIGIN3 == U_ORIGIN3 {
			err, _ = message.readCoord()
		}
		if bits&U_ANGLE3 == U_ANGLE3 {
			err, _ = message.readAngle()
		}
	}
	return nil
}

func (message *Message) Svc_packetentities(mvd *Mvd) error {
	for {
		err, w := message.readShort()
		if err != nil {
			return err
		}
		if w == 0 {
			break
		}

		w &= ^511
		bits := w

		if bits&U_MOREBITS == U_MOREBITS {
			err, i := message.readByte()
			if err != nil {
				return err
			}
			bits |= int(i)
		}

		if bits&U_MODEL == U_MODEL {
			err, _ = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_FRAME == U_FRAME {
			err, _ = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_COLORMAP == U_COLORMAP {
			err, _ = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_SKIN == U_SKIN {
			err, _ = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_EFFECTS == U_EFFECTS {
			err, _ = message.readByte()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN1 == U_ORIGIN1 {
			err, _ = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE1 == U_ANGLE1 {
			err, _ = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN2 == U_ORIGIN2 {
			err, _ = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE2 == U_ANGLE2 {
			err, _ = message.readAngle()
			if err != nil {
				return err
			}
		}
		if bits&U_ORIGIN3 == U_ORIGIN3 {
			err, _ = message.readCoord()
			if err != nil {
				return err
			}
		}
		if bits&U_ANGLE3 == U_ANGLE3 {
			err, _ = message.readAngle()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (message *Message) Svc_temp_entity(mvd *Mvd) error {

	err, t := message.readByte()
	if err != nil {
		return err
	}

	if t == TE_GUNSHOT || t == TE_BLOOD {
		err, _ = message.readByte()
		if err != nil {
			return err
		}
	}

	if t == TE_LIGHTNING1 || t == TE_LIGHTNING2 || t == TE_LIGHTNING3 {
		err, _ = message.readShort()
		if err != nil {
			return err
		}
		err, _ = message.readCoord()
		if err != nil {
			return err
		}
		err, _ = message.readCoord()
		if err != nil {
			return err
		}
		err, _ = message.readCoord()
		if err != nil {
			return err
		}
	}

	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_print(mvd *Mvd) error {
	err, from := message.readByte()
	if err != nil {
		return err
	}
	err, s := message.readString()
	if err != nil {
		return err
	}
	mvd.State.Messages = append(mvd.State.Messages, ServerMessage{int(from), s})
	return nil
}

func (message *Message) Svc_serverinfo(mvd *Mvd) error {
	err, key := message.readString()
	if err != nil {
		return err
	}
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
	err, s := message.readString()
	if err != nil {
		return err
	}
	mvdPrint(s)
	return nil
}

func (message *Message) Svc_setinfo(mvd *Mvd) error {
	err, pnum := message.readByte() // num
	if err != nil {
		return err
	}
	err, key := message.readString() // key
	if err != nil {
		return err
	}
	err, value := message.readString() // value
	if err != nil {
		return err
	}
	mvd.State.Players[pnum].Setinfo[key] = value
	mvd.emitEventPlayer(&mvd.State.Players[int(pnum)], pnum, PE_USERINFO)
	return nil
}

func (message *Message) Svc_damage(mvd *Mvd) error {
	err, _ := message.readByte() // armor
	if err != nil {
		return err
	}
	err, _ = message.readByte() // blood
	if err != nil {
		return err
	}
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_chokecount(mvd *Mvd) error {
	err, _ := message.readByte()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_spawnstatic(mvd *Mvd) error {
	err, _ := message.readByte()
	if err != nil {
		return err
	}
	err, _ = message.readByte()
	if err != nil {
		return err
	}
	err, _ = message.readByte()
	if err != nil {
		return err
	}
	err, _ = message.readByte()
	if err != nil {
		return err
	}
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
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
	err, _ := message.readShort()
	if err != nil {
		return err
	}
	return nil
}

func (message *Message) Svc_intermission(mvd *Mvd) error {
	err, _ := message.readCoord()
	if err != nil {
		return err
	}
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	err, _ = message.readCoord()
	if err != nil {
		return err
	}
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
	err, _ = message.readAngle()
	if err != nil {
		return err
	}
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

func (message *Message) readBytes(count uint) (error, *bytes.Buffer) {
	if message.offset+count > message.size {
		return errors.New("reading beyong message length"), nil
	}
	b := bytes.NewBuffer(message.data[message.offset : message.offset+count])
	message.offset += count
	return nil, b
}

func (message *Message) readByte() (error, byte) {
	var b byte
	err, barray := message.readBytes(1)
	if err != nil {
		return err, byte(0)
	}
	err = binary.Read(barray, binary.BigEndian, &b)
	if err != nil {
		return err, byte(0)
	}
	return nil, b
}

func (message *Message) readLong() (error, int) {
	var i int32
	err, b := message.readBytes(4)
	if err != nil {
		return err, 0
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}
	return nil, int(i)
}

func (message *Message) readFloat() (error, float32) {
	var i float32
	err, b := message.readBytes(4)
	if err != nil {
		return err, 0
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}
	return nil, float32(i)
}

func (message *Message) readString() (error, string) {
	b := make([]byte, 0)
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
	return nil, string(b)
}

func (message *Message) readCoord() (error, float32) {
	if message.mvd.demo.fte_pext&FTE_PEXT_FLOATCOORDS == FTE_PEXT_FLOATCOORDS {
		err, f := message.readFloat()
		if err != nil {
			return err, 0
		}
		return nil, f
	}
	err, b := message.readShort()
	if err != nil {
		return err, 0
	}
	return nil, float32(b) * (1.0 / 8)
}

func (message *Message) readAngle() (error, float32) {
	if message.mvd.demo.fte_pext&FTE_PEXT_FLOATCOORDS == FTE_PEXT_FLOATCOORDS {

		err, a := message.readAngle16()
		if err != nil {
			return err, 0

		}
		return nil, a
	}
	err, b := message.readByte()
	if err != nil {
		return err, 0
	}
	return nil, float32(b) * (360.0 / 256.0)
}

func (message *Message) readAngle16() (error, float32) {
	err, b := message.readShort()
	if err != nil {
		return err, 0
	}

	return nil, float32(b) * (360.0 / 65536)
}

func (message *Message) readShort() (error, int) {
	var i int16
	err, b := message.readBytes(2)
	if err != nil {
		return err, 0
	}
	err = binary.Read(b, binary.LittleEndian, &i)
	if err != nil {
		return err, 0
	}
	return nil, int(i)
}
