package mvdreader

// protocol frame type
//go:generate stringer -type=DEM_TYPE
type DEM_TYPE uint

const (
	dem_cmd DEM_TYPE = iota
	dem_read
	dem_set
	dem_multiple
	dem_single
	dem_stats
	dem_all
)

//server command types
//go:generate stringer -type=SVC_TYPE
type SVC_TYPE uint

const (
	svc_bad                 SVC_TYPE = 0
	svc_nop                 SVC_TYPE = 1
	svc_disconnect          SVC_TYPE = 2
	svc_updatestat          SVC_TYPE = 3  // [byte] [byte]
	nq_svc_version          SVC_TYPE = 4  // [long] server version
	svc_setview             SVC_TYPE = 5  // [short] entity number
	svc_sound               SVC_TYPE = 6  // <see code>
	nq_svc_time             SVC_TYPE = 7  // [float] server time
	svc_print               SVC_TYPE = 8  // [byte] id [string] null terminated string
	svc_stufftext           SVC_TYPE = 9  // [string] stuffed into client's console buffer
	svc_setangle            SVC_TYPE = 10 // [angle3] set the view angle to this absolute value
	svc_serverdata          SVC_TYPE = 11 // [long] protocol ...
	svc_lightstyle          SVC_TYPE = 12 // [byte] [string]
	nq_svc_updatename       SVC_TYPE = 13 // [byte] [string]
	svc_updatefrags         SVC_TYPE = 14 // [byte] [short]
	nq_svc_clientdata       SVC_TYPE = 15 // <shortbits + data>
	svc_stopsound           SVC_TYPE = 16 // <see code>
	nq_svc_updatecolors     SVC_TYPE = 17 // [byte] [byte] [byte]
	nq_svc_particle         SVC_TYPE = 18 // [vec3] <variable>
	svc_damage              SVC_TYPE = 19
	svc_spawnstatic         SVC_TYPE = 20
	svc_spawnbinary         SVC_TYPE = 21
	svc_spawnbaseline       SVC_TYPE = 22
	svc_temp_entity         SVC_TYPE = 23 // variable
	svc_setpause            SVC_TYPE = 24 // [byte] on / off
	nq_svc_signonnum        SVC_TYPE = 25 // [byte]  used for the signon sequence
	svc_centerprint         SVC_TYPE = 26 // [string] to put in center of the screen
	svc_killedmonster       SVC_TYPE = 27
	svc_foundsecret         SVC_TYPE = 28
	svc_spawnstaticsound    SVC_TYPE = 29 // [coord3] [byte] samp [byte] vol [byte] aten
	svc_intermission        SVC_TYPE = 30 // [vec3_t] origin [vec3_t] angle
	svc_finale              SVC_TYPE = 31 // [string] text
	svc_cdtrack             SVC_TYPE = 32 // [byte] track
	svc_sellscreen          SVC_TYPE = 33
	nq_svc_cutscene         SVC_TYPE = 34 // same as svc_smallkick
	svc_smallkick           SVC_TYPE = 34 // set client punchangle to 2
	svc_bigkick             SVC_TYPE = 35 // set client punchangle to 4
	svc_updateping          SVC_TYPE = 36 // [byte] [short]
	svc_updateentertime     SVC_TYPE = 37 // [byte] [float]
	svc_updatestatlong      SVC_TYPE = 38 // [byte] [long]
	svc_muzzleflash         SVC_TYPE = 39 // [short] entity
	svc_updateuserinfo      SVC_TYPE = 40 // [byte] slot [long] uid [string] userinfo
	svc_download            SVC_TYPE = 41 // [short] size [size bytes]
	svc_playerinfo          SVC_TYPE = 42 // variable
	svc_nails               SVC_TYPE = 43 // [byte] num [48 bits] xyzpy 12 12 12 4 8
	svc_chokecount          SVC_TYPE = 44 // [byte] packets choked
	svc_modellist           SVC_TYPE = 45 // [strings]
	svc_soundlist           SVC_TYPE = 46 // [strings]
	svc_packetentities      SVC_TYPE = 47 // [...]
	svc_deltapacketentities SVC_TYPE = 48 // [...]
	svc_maxspeed            SVC_TYPE = 49 // maxspeed change, for prediction
	svc_entgravity          SVC_TYPE = 50 // gravity change, for prediction
	svc_setinfo             SVC_TYPE = 51 // setinfo on a client
	svc_serverinfo          SVC_TYPE = 52 // serverinfo
	svc_updatepl            SVC_TYPE = 53 // [byte] [byte]
	svc_nails2              SVC_TYPE = 54
	svc_fte_modellistshort  SVC_TYPE = 60
	svc_fte_spawnbaseline2  SVC_TYPE = 66
	svc_qizmovoice          SVC_TYPE = 83
)

//go:generate stringer -type=PROTOCOL_VERSION
type PROTOCOL_VERSION uint32

const (
	protocol_standard PROTOCOL_VERSION = 28
	protocol_fte      PROTOCOL_VERSION = (('F' << 0) + ('T' << 8) + ('E' << 16) + ('X' << 24)) //fte extensions.
	protocol_fte2     PROTOCOL_VERSION = (('F' << 0) + ('T' << 8) + ('E' << 16) + ('2' << 24)) //fte extensions.
	protocol_mvd1     PROTOCOL_VERSION = (('M' << 0) + ('V' << 8) + ('D' << 16) + ('1' << 24)) //mvdsv extensions
)

//go:generate stringer -type=FTE_PROTOCOL_EXTENSION
type FTE_PROTOCOL_EXTENSION uint

const (
	FTE_PEXT_TRANS             FTE_PROTOCOL_EXTENSION = 0x00000008 // .alpha support
	FTE_PEXT_ACCURATETIMINGS   FTE_PROTOCOL_EXTENSION = 0x00000040
	FTE_PEXT_HLBSP             FTE_PROTOCOL_EXTENSION = 0x00000200 //stops fte servers from complaining
	FTE_PEXT_MODELDBL          FTE_PROTOCOL_EXTENSION = 0x00001000
	FTE_PEXT_ENTITYDBL         FTE_PROTOCOL_EXTENSION = 0x00002000 //max =of 1024 ents instead of 512
	FTE_PEXT_ENTITYDBL2        FTE_PROTOCOL_EXTENSION = 0x00004000 //max =of 1024 ents instead of 512
	FTE_PEXT_FLOATCOORDS       FTE_PROTOCOL_EXTENSION = 0x00008000 //supports =floating point origins.
	FTE_PEXT_SPAWNSTATIC2      FTE_PROTOCOL_EXTENSION = 0x00400000 //Sends =an entity delta instead of a baseline.
	FTE_PEXT_256PACKETENTITIES FTE_PROTOCOL_EXTENSION = 0x01000000 //Client =can recieve 256 packet entities.
	FTE_PEXT_CHUNKEDDOWNLOADS  FTE_PROTOCOL_EXTENSION = 0x20000000 //alternate =file download method. Hopefully it'll give quadroupled download speed, especially on higher pings.
)

//go:generate stringer -type=MVD_PROTOCOL_EXTENSION
type MVD_PROTOCOL_EXTENSION uint

const (
	MVD_PEXT1_FLOATCOORDS     MVD_PROTOCOL_EXTENSION = 0x00000001 // FTE_PEXT_FLOATCOORDS but for entity/player coords only
	MVD_PEXT1_HIGHLAGTELEPORT MVD_PROTOCOL_EXTENSION = 0x00000002 // Adjust movement direction for frames following teleport
)

//go:generate stringer -type=DF_TYPE
type DF_TYPE uint

const (
	DF_ORIGIN      DF_TYPE = 1
	DF_ANGLES      DF_TYPE = (1 << 3)
	DF_EFFECTS     DF_TYPE = (1 << 6)
	DF_SKINNUM     DF_TYPE = (1 << 7)
	DF_DEAD        DF_TYPE = (1 << 8)
	DF_GIB         DF_TYPE = (1 << 9)
	DF_WEAPONFRAME DF_TYPE = (1 << 10)
	DF_MODEL       DF_TYPE = (1 << 11)
)

//go:generate stringer -type=SND_TYPE
type SND_TYPE uint

const (
	SND_VOLUME      SND_TYPE = (1 << 15)
	SND_ATTENUATION SND_TYPE = (1 << 14)
)

const (
	int_size       = 4
	float_size     = 4
	vec3_t_size    = (3 * float_size)
	short_size     = 2
	byte_size      = 1
	userCmd_t_size = (12 + 3*byte_size + 3*short_size + 1*vec3_t_size)
)

const (
	U_ORIGIN1  = (1 << 9)
	U_ORIGIN2  = (1 << 10)
	U_ORIGIN3  = (1 << 11)
	U_ANGLE2   = (1 << 12)
	U_FRAME    = (1 << 13)
	U_REMOVE   = (1 << 14) // REMOVE this entity, don't add it
	U_MOREBITS = (1 << 15)
	// if MOREBITS is set, these additional flags are read in next
	U_ANGLE1   = (1 << 0)
	U_ANGLE3   = (1 << 1)
	U_MODEL    = (1 << 2)
	U_COLORMAP = (1 << 3)
	U_SKIN     = (1 << 4)
	U_EFFECTS  = (1 << 5)
	U_SOLID    = (1 << 6) // the entity should be solid for prediction
)

// FTE
const (
	U_FTE_EVENVENMORE = (1 << 7)
	U_FTE_YETMORE     = (1 << 7)
	U_FTE_TRANS       = (1 << 1) //transparency value
	U_FTE_ENTITYDBL   = (1 << 5) //use an extra byte for origin parts, cos one of them is off
	U_FTE_ENTITYDBL2  = (1 << 6) //use an extra byte for origin parts, cos one of them is off
)

//go:generate stringer -type=TE_TYPE
type TE_TYPE byte

const (
	TE_SPIKE          TE_TYPE = 0
	TE_SUPERSPIKE     TE_TYPE = 1
	TE_GUNSHOT        TE_TYPE = 2
	TE_EXPLOSION      TE_TYPE = 3
	TE_TAREXPLOSION   TE_TYPE = 4
	TE_LIGHTNING1     TE_TYPE = 5
	TE_LIGHTNING2     TE_TYPE = 6
	TE_WIZSPIKE       TE_TYPE = 7
	TE_KNIGHTSPIKE    TE_TYPE = 8
	TE_LIGHTNING3     TE_TYPE = 9
	TE_LAVASPLASH     TE_TYPE = 10
	TE_TELEPORT       TE_TYPE = 11
	TE_BLOOD          TE_TYPE = 12
	TE_LIGHTNINGBLOOD TE_TYPE = 13
)

//go:generate stringer -type=STAT_TYPE
type STAT_TYPE int

const (
	STAT_HEALTH        STAT_TYPE = 0
	STAT_FRAGS         STAT_TYPE = 1
	STAT_WEAPON        STAT_TYPE = 2
	STAT_AMMO          STAT_TYPE = 3
	STAT_ARMOR         STAT_TYPE = 4
	STAT_WEAPONFRAME   STAT_TYPE = 5
	STAT_SHELLS        STAT_TYPE = 6
	STAT_NAILS         STAT_TYPE = 7
	STAT_ROCKETS       STAT_TYPE = 8
	STAT_CELLS         STAT_TYPE = 9
	STAT_ACTIVEWEAPON  STAT_TYPE = 10
	STAT_TOTALSECRETS  STAT_TYPE = 11
	STAT_TOTALMONSTERS STAT_TYPE = 12
	STAT_SECRETS       STAT_TYPE = 13 // bumped on client side by svc_foundsecret
	STAT_MONSTERS      STAT_TYPE = 14 // bumped by svc_killedmonster
	STAT_ITEMS         STAT_TYPE = 15
	STAT_VIEWHEIGHT    STAT_TYPE = 16 // Z_EXT_VIEWHEIGHT protocol extension
	STAT_TIME          STAT_TYPE = 17 // Z_EXT_TIME extension
	STAT_WTF           STAT_TYPE = 18 // Z_EXT_TIME extension
)

//go:generate stringer -type=IT_TYPE
type IT_TYPE int

const (
	IT_SHOTGUN          IT_TYPE = 1
	IT_SUPER_SHOTGUN    IT_TYPE = 2
	IT_NAILGUN          IT_TYPE = 4
	IT_SUPER_NAILGUN    IT_TYPE = 8
	IT_GRENADE_LAUNCHER IT_TYPE = 16
	IT_ROCKET_LAUNCHER  IT_TYPE = 32
	IT_LIGHTNING        IT_TYPE = 64
	IT_SUPER_LIGHTNING  IT_TYPE = 128
	IT_SHELLS           IT_TYPE = 256
	IT_NAILS            IT_TYPE = 512
	IT_ROCKETS          IT_TYPE = 1024
	IT_CELLS            IT_TYPE = 2048
	IT_AXE              IT_TYPE = 4096
	IT_ARMOR1           IT_TYPE = 8192
	IT_ARMOR2           IT_TYPE = 16384
	IT_ARMOR3           IT_TYPE = 32768
	IT_SUPERHEALTH      IT_TYPE = 65536
	IT_KEY1             IT_TYPE = 131072
	IT_KEY2             IT_TYPE = 262144
	IT_INVISIBILITY     IT_TYPE = 524288
	IT_INVULNERABILITY  IT_TYPE = 1048576
	IT_SUIT             IT_TYPE = 2097152
	IT_QUAD             IT_TYPE = 4194304
	IT_UNKNOWN1         IT_TYPE = (1 << 23)
	IT_UNKNOWN2         IT_TYPE = (1 << 24)
	IT_UNKNOWN3         IT_TYPE = (1 << 25)
	IT_UNKNOWN4         IT_TYPE = (1 << 26)
	IT_UNKNOWN5         IT_TYPE = (1 << 27)
	IT_SIGIL1           IT_TYPE = (1 << 28)
	IT_SIGIL2           IT_TYPE = (1 << 29)
	IT_SIGIL3           IT_TYPE = (1 << 30)
	IT_SIGIL4           IT_TYPE = (1 << 31)
)
