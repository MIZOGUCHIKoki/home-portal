package model

// seed用
type MethodSeed struct {
	Identifier string
	MethodName string
}

// DB用
type Method struct {
	MethodID   int64
	Identifier string
	MethodName string
}

var DefaultMethods = []MethodSeed{
	{
		Identifier: "cash",
		MethodName: "現金",
	},
	{
		Identifier: "jal_jcb",
		MethodName: "JAL Card (JCB)",
	},
	{
		Identifier: "jal_master",
		MethodName: "JAL Card (Master)",
	},
	{
		Identifier: "jal_amex",
		MethodName: "JAL Card (AMEX)",
	},
	{
		Identifier: "bic_camera_suica_jcb",
		MethodName: "BIC CAMERA Suica (JCB)",
	},
	{
		Identifier: "0001",
		MethodName: "みずほ銀行",
	},
	{
		Identifier: "0038",
		MethodName: "ドコモSMTBネット銀行",
	},
}
