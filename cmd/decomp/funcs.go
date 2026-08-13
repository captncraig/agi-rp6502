package main

type argType byte

const (
	argVar argType = iota
	argNum
	argFlag
	argCtrl
	argMsg
	argObj
	argItem
	argString
	argColor
)

type fDef struct {
	name string
	args []argType
}

var functions = map[byte]fDef{
	0x09: {
		name: "lindirectv",
		args: []argType{argVar, argVar},
	},
	0x0a: {
		name: "rindirect",
		args: []argType{argVar, argVar},
	},
	0x0b: {
		name: "lindirectn",
		args: []argType{argVar, argNum},
	},
	0x0c: {
		name: "set",
		args: []argType{argFlag},
	},
	0x0d: {
		name: "reset",
		args: []argType{argFlag},
	},
	0x0e: {
		name: "toggle",
		args: []argType{argFlag},
	},
	0x0f: {
		name: "set",
		args: []argType{argVar},
	},
	0x10: {
		name: "reset",
		args: []argType{argVar},
	},
	// 0x11: {
	// 	name: "reset",
	// 	args: []argType{argVar},
	// },
	0x12: {
		name: "new.room",
		args: []argType{argNum},
	},
	0x13: {
		name: "new.room",
		args: []argType{argVar},
	},
	0x14: {
		name: "load.logic",
		args: []argType{argNum},
	},
	0x15: {
		name: "load.logic",
		args: []argType{argVar},
	},
	0x16: {
		name: "call",
		args: []argType{argNum},
	},
	0x17: {
		name: "call",
		args: []argType{argVar},
	},
	0x18: {
		name: "load.pic",
		args: []argType{argVar},
	},
	0x19: {
		name: "draw.pic",
		args: []argType{argVar},
	},
	0x1a: {
		name: "show.pic",
		args: []argType{},
	},
	0x1b: {
		name: "discard.pic",
		args: []argType{argVar},
	},
	0x1c: {
		name: "overlay.pic",
		args: []argType{argVar},
	},
	0x1d: {
		name: "show.pri.screen",
		args: []argType{},
	},
	0x1e: {
		name: "load.view",
		args: []argType{argNum},
	},
	0x1f: {
		name: "load.view",
		args: []argType{argVar},
	},
	0x20: {
		name: "discard.view",
		args: []argType{argNum},
	},
	0x21: {
		name: "animate.obj",
		args: []argType{argObj},
	},
	0x22: {
		name: "unanimate.all",
		args: []argType{},
	},
	0x23: {
		name: "draw",
		args: []argType{argObj},
	},
	0x24: {
		name: "erase",
		args: []argType{argObj},
	},
	0x25: {
		name: "position",
		args: []argType{argObj, argNum, argNum},
	},
	0x26: {
		name: "position",
		args: []argType{argObj, argNum, argVar},
	},
	0x27: {
		name: "get.posn",
		args: []argType{argObj, argVar, argVar},
	},
	0x28: {
		name: "repositon",
		args: []argType{argObj, argVar, argVar},
	},
	0x29: {
		name: "set.view",
		args: []argType{argObj, argNum},
	},
	0x2a: {
		name: "set.view",
		args: []argType{argObj, argVar},
	},
	0x2b: {
		name: "set.loop",
		args: []argType{argObj, argNum},
	},
	0x2c: {
		name: "set.loop",
		args: []argType{argObj, argVar},
	},
	0x2d: {
		name: "fix.loop",
		args: []argType{argObj},
	},
	0x2e: {
		name: "release.loop",
		args: []argType{argObj},
	},
	0x2f: {
		name: "set.cel",
		args: []argType{argObj, argNum},
	},
	0x30: {
		name: "set.cel",
		args: []argType{argObj, argVar},
	},
	0x31: {
		name: "last.cel",
		args: []argType{argObj, argVar},
	},
	0x32: {
		name: "current.cel",
		args: []argType{argObj, argVar},
	},
	0x33: {
		name: "current.loop",
		args: []argType{argObj, argVar},
	},
	0x34: {
		name: "current.view",
		args: []argType{argObj, argVar},
	},
	0x35: {
		name: "number.of.loops",
		args: []argType{argObj, argVar},
	},
	0x36: {
		name: "set.priority",
		args: []argType{argObj, argNum},
	},
	0x37: {
		name: "set.priority",
		args: []argType{argObj, argVar},
	},
	0x38: {
		name: "release.priority",
		args: []argType{argObj},
	},
	0x39: {
		name: "get.priority",
		args: []argType{argObj, argVar},
	},
	0x3a: {
		name: "stop.update",
		args: []argType{argObj},
	},
	0x3b: {
		name: "start.update",
		args: []argType{argObj},
	},
	0x3c: {
		name: "force.update",
		args: []argType{argObj},
	},
	0x3d: {
		name: "ignore.horizon",
		args: []argType{argObj},
	},
	0x3e: {
		name: "observe.horizon",
		args: []argType{argObj},
	},
	0x3f: {
		name: "set.horizon",
		args: []argType{argNum},
	},
	0x40: {
		name: "object.on.water",
		args: []argType{argObj},
	},
	0x41: {
		name: "object.on.land",
		args: []argType{argObj},
	},
	0x42: {
		name: "object.on.anything",
		args: []argType{argObj},
	},
	0x43: {
		name: "ignore.objs",
		args: []argType{argObj},
	},
	0x44: {
		name: "observe.objs",
		args: []argType{argObj},
	},
	0x45: {
		name: "distance",
		args: []argType{argObj, argObj, argVar},
	},
	0x46: {
		name: "stop.cycling",
		args: []argType{argObj},
	},
	0x47: {
		name: "start.cycling",
		args: []argType{argObj},
	},
	0x48: {
		name: "normal.cycle",
		args: []argType{argObj},
	},
	0x49: {
		name: "end.of.loop",
		args: []argType{argObj, argFlag},
	},
	0x4a: {
		name: "reverse.cycle",
		args: []argType{argObj},
	},
	0x4b: {
		name: "reverse.loop",
		args: []argType{argObj, argFlag},
	},
	0x4c: {
		name: "cycle.time",
		args: []argType{argObj, argVar},
	},
	0x4d: {
		name: "stop.motion",
		args: []argType{argObj},
	},
	0x4e: {
		name: "start.motion",
		args: []argType{argObj},
	},
	0x4f: {
		name: "step.size",
		args: []argType{argObj, argVar},
	},
	0x50: {
		name: "step.time",
		args: []argType{argObj, argNum},
	},
	0x51: {
		name: "move.obj",
		args: []argType{argObj, argNum, argNum, argNum, argFlag},
	},
	0x52: {
		name: "move.obj",
		args: []argType{argObj, argVar, argVar, argNum, argFlag},
	},
	0x53: {
		name: "follow.ego",
		args: []argType{argObj, argNum, argFlag},
	},
	0x54: {
		name: "wander",
		args: []argType{argObj},
	},
	0x55: {
		name: "normal.motion",
		args: []argType{argObj},
	},
	0x56: {
		name: "set.dir",
		args: []argType{argObj, argVar},
	},
	0x57: {
		name: "	get.dir",
		args: []argType{argObj, argVar},
	},
	0x58: {
		name: "ignore.blocks",
		args: []argType{argObj},
	},
	0x59: {
		name: "observe.blocks",
		args: []argType{argObj},
	},
	0x5a: {
		name: "block",
		args: []argType{argNum, argNum, argNum, argNum},
	},
	0x5b: {
		name: "unblock",
		args: []argType{},
	},
	0x5c: {
		name: "get",
		args: []argType{argItem},
	},
	0x5d: {
		name: "get",
		args: []argType{argVar},
	},
	0x5e: {
		name: "drop",
		args: []argType{argItem},
	},
	0x5f: {
		name: "put",
		args: []argType{argItem, argNum},
	},
	0x60: {
		name: "put",
		args: []argType{argVar, argVar},
	},
	0x61: {
		name: "get.room",
		args: []argType{argVar, argVar},
	},
	0x62: {
		name: "load.sound",
		args: []argType{argNum},
	},
	0x63: {
		name: "sound",
		args: []argType{argNum, argFlag},
	},
	0x64: {
		name: "stop.sound",
		args: []argType{},
	},
	0x65: {
		name: "print",
		args: []argType{argMsg},
	},
	0x66: {
		name: "print",
		args: []argType{argVar},
	},
	0x67: {
		name: "display",
		args: []argType{argNum, argNum, argMsg},
	},
	0x68: {
		name: "display",
		args: []argType{argVar, argVar, argVar},
	},
	0x69: {
		name: "clear.lines",
		args: []argType{argNum, argNum, argColor},
	},
	0x6a: {
		name: "text.screen",
		args: []argType{},
	},
	0x6b: {
		name: "graphics",
		args: []argType{},
	},
	0x6c: {
		name: "set.cursor.char",
		args: []argType{argMsg},
	},
	0x6d: {
		name: "set.text.attribute",
		args: []argType{argNum, argNum},
	},
	0x6e: {
		name: "shake.screen",
		args: []argType{argNum},
	},
	0x6f: {
		name: "configure.screen",
		args: []argType{argNum, argNum, argNum},
	},
	0x70: {
		name: "status.line.on",
		args: []argType{},
	},
	0x71: {
		name: "status.line.off",
		args: []argType{},
	},
	0x72: {
		name: "set.string",
		args: []argType{argString, argMsg},
	},
	0x73: {
		name: "get.string",
		args: []argType{argString, argMsg, argNum, argNum, argNum},
	},
	// 0x74: {
	// 	name: "",
	// 	args: []argType{},
	// },
	0x75: {
		name: "parse",
		args: []argType{argString},
	},
	0x76: {
		name: "get.num",
		args: []argType{argString, argVar},
	},
	0x77: {
		name: "prevent.input",
		args: []argType{},
	},
	0x78: {
		name: "accept.input",
		args: []argType{},
	},
	0x79: {
		// todo: maybe special arg types for keys
		name: "set.key",
		args: []argType{argNum, argNum, argNum},
	},
	0x7a: {
		name: "add.to.pic",
		args: []argType{argNum, argNum, argNum, argNum, argNum, argNum, argNum},
	},
	0x7b: {
		name: "add.to.pic",
		args: []argType{argVar, argVar, argVar, argVar, argVar, argVar, argVar},
	},
	0x7c: {
		name: "status",
		args: []argType{},
	},
	0x7d: {
		name: "save.game",
		args: []argType{},
	},
	0x7e: {
		name: "restore.game",
		args: []argType{},
	},
	// 0x7f: {
	// 	name: "",
	// 	args: []argType{},
	// },
	0x80: {
		name: "restart.game",
		args: []argType{},
	},
	0x81: {
		name: "show.obj",
		args: []argType{argNum},
	},
	0x82: {
		name: "random",
		args: []argType{argNum, argNum, argVar},
	},
	0x83: {
		name: "program.control",
		args: []argType{},
	},
	0x84: {
		name: "player.control",
		args: []argType{},
	},
	0x85: {
		name: "obj.status.v",
		args: []argType{argVar},
	},
	0x86: {
		name: "quit",
		args: []argType{argNum},
	},
	0x87: {
		name: "show.mem",
		args: []argType{},
	},
	0x88: {
		name: "pause",
		args: []argType{},
	},
	0x89: {
		name: "echo.line",
		args: []argType{},
	},
	0x8a: {
		name: "cancel.line",
		args: []argType{},
	},
	0x8b: {
		name: "init.joy",
		args: []argType{},
	},
	0x8c: {
		name: "toggle.monitor",
		args: []argType{},
	},
	0x8d: {
		name: "version",
		args: []argType{},
	},
	0x8e: {
		name: "script.size",
		args: []argType{argNum},
	},
	0x8f: {
		name: "set.game.id",
		args: []argType{argMsg},
	},
	0x90: {
		name: "log",
		args: []argType{argMsg},
	},
	0x91: {
		name: "set.scan.start",
		args: []argType{},
	},
	0x92: {
		name: "reset.scan.start",
		args: []argType{},
	},
	0x93: {
		name: "reposition.to",
		args: []argType{argObj, argNum, argNum},
	},
	0x94: {
		name: "reposition.to",
		args: []argType{argObj, argVar, argVar},
	},
	// 0x95: {
	// 	name: "",
	// 	args: []argType{},
	// },
	0x96: {
		name: "trace.info",
		args: []argType{argNum, argNum, argNum},
	},
	0x97: {
		name: "print.at",
		args: []argType{argMsg, argNum, argNum, argNum},
	},
	0x98: {
		name: "print.at",
		args: []argType{argVar, argNum, argNum, argNum},
	},
	// 0x99: {
	// 	name: "",
	// 	args: []argType{},
	// },
	0x9a: {
		name: "clear.text.rect",
		args: []argType{argNum, argNum, argNum, argNum, argNum},
	},
	// 0x9b: {
	// 	name: "",
	// 	args: []argType{},
	// },
	0x9c: {
		name: "set.menu",
		args: []argType{argMsg},
	},
	0x9d: {
		name: "set.menu.member",
		args: []argType{argMsg, argCtrl},
	},
	0x9e: {
		name: "submit.menu",
		args: []argType{},
	},
	0x9f: {
		name: "enable.member",
		args: []argType{argCtrl},
	},
	0xa0: {
		name: "disable.member",
		args: []argType{argCtrl},
	},
	0xa1: {
		name: "menu.input",
		args: []argType{},
	},
	0xa2: {
		name: "show.obj",
		args: []argType{argVar},
	},
	0xa3: {
		name: "open.dialogue",
		args: []argType{},
	},
	0xa4: {
		name: "close.dialogue",
		args: []argType{},
	},
	0xa5: {
		name: "mul",
		args: []argType{argVar, argNum},
	},
	0xa6: {
		name: "mul",
		args: []argType{argVar, argVar},
	},
	// 0xa7: {
	// 	name: "",
	// 	args: []argType{},
	// },
	// 0xa8: {
	// 	name: "",
	// 	args: []argType{},
	// },
	0xa9: {
		name: "close.window",
		args: []argType{},
	},
	0xaa: {
		name: "set.simple",
		args: []argType{argNum},
	},
	// 0xab: {
	// 	name: "",
	// 	args: []argType{},
	// },
	// 0xac: {
	// 	name: "",
	// 	args: []argType{},
	// },
	0xad: {
		name: "hold.key",
		args: []argType{},
	},
	// 0xae: {
	// 	name: "",
	// 	args: []argType{},
	// },
	// 0xaf: {
	// 	name: "",
	// 	args: []argType{},
	// },
	// 0xb0: {
	// 	name: "",
	// 	args: []argType{},
	// },
	// 0xb1: {
	// 	name: "",
	// 	args: []argType{},
	// },
	// 0xb2: {
	// 	name: "",
	// 	args: []argType{},
	// },
	// 0xb3: {
	// 	name: "",
	// 	args: []argType{},
	// },
	// 0xb4: {
	// 	name: "",
	// 	args: []argType{},
	// },
	// 0xb5: {
	// 	name: "",
	// 	args: []argType{},
	// },
	// 0xb6: {
	// 	name: "",
	// 	args: []argType{},
	// },
}

var testFunctions = map[byte]fDef{
	0x09: {
		name: "has",
		args: []argType{argItem},
	},
	0x0a: {
		name: "obj.in.room",
		args: []argType{argItem, argVar},
	},
	0x0b: {
		name: "posn",
		args: []argType{argObj, argNum, argNum, argNum, argNum},
	},
	0x0c: {
		name: "controller",
		args: []argType{argCtrl},
	},
	0x0d: {
		name: "have.key",
		args: []argType{},
	},
	// 0x0e: {
	// 	name: "",
	// 	args: []argType{},
	// },
	0x0f: {
		name: "compare.strings",
		args: []argType{argString, argString},
	},
	0x10: {
		name: "obj.in.box",
		args: []argType{argObj, argNum, argNum, argNum, argNum},
	},
	0x11: {
		name: "center.posn",
		args: []argType{argObj, argNum, argNum, argNum, argNum},
	},
	0x12: {
		name: "right.posn",
		args: []argType{argObj, argNum, argNum, argNum, argNum},
	},
}
