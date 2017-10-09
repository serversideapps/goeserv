package main

var(

TRANSLATIONS map[string]string = map[string]string{		
		"presentation.engine":"Engine",
		"presentation.hybernate":"Hybernate",
		"presentation.nothybernated":"Not hybernated",
		"presentation.hybernated":"Hybernated",
		"presentation.changehybernated":"Change",
		"presentation.archive":"Archive",
		"presentation.notarchived":"Not archived",
		"presentation.archived":"Archived",
		"presentation.changearchived":"Change",
		"presentation.access.denied":"Access denied!",
		"presentation.deleted":"Presentation deleted.",
		"presentation.doesnotexist":"Presentation does not exist!",
		"ccfg.change":"Change",
		"ccfg.essay":"Essay",
		"ccfg.saveessay":"Save essay",
		"ccfg.savemovenote":"Save",
		"ccfg.editmovenote":"Edit",
		"ccfg.nodelink":"Node link",
		"ccfg.nodeurl":"Node url",
		"ccfg.version":"Version",
		"ccfg.white":"White",
		"ccfg.black":"Black",
		"ccfg.yellow":"Yellow",
		"ccfg.red":"Red",
		"ccfg.variant":"Variant",
		"ccfg.timecontrol":"Time control",
		"ccfg.gamesanalysis":"Game analysis",
		"ccfg.whiteresigned":"White resigned",
		"ccfg.blackresigned":"Black resigned",
		"ccfg.whiteflagged":"White flagged",
		"ccfg.blackflagged":"Black flagged",
		"ccfg.whitemated":"White mated",
		"ccfg.blackmated":"Black mated",
		"ccfg.open":"Open",
		"ccfg.inprogress":"In progress",
		"ccfg.terminated":"Terminated",
		"ccfg.result":"Result",
		"ccfg.created":"Created",
		"ccfg.status":"Status",
		"ccfg.refresh":"Refresh",
		"ccfg.load":"Load",
		"ccfg.loadingboard":"Loading board ...",
		"ccfg.importpgn":"Import PGN : ",
		"ccfg.savenotes":"Save notes",
		"ccfg.upload":"Upload",
		"ccfg.title":"Title",
		"ccfg.owner":"Owner",
		"ccfg.id":"Id",
		"ccfg.gennew":"Generate new",
		"ccfg.resign":"Resign",
		"ccfg.stand":"Stand",
		"ccfg.play":"Play",
		"ccfg.playai":"Play AI",
		"ccfg.flip":"Flip",
		"ccfg.yes":"Yes",
		"ccfg.no":"No",
		"ccfg.current":"Current",
		"ccfg.book":"Book",
		"ccfg.notes":"Notes",
		"ccfg.presentation":"Presentation",
		"ccfg.connecting":"Connecting ...",
		"ccfg.createtable":"Create table",
		"ccfg.creatingtable":"Creating table",
		"ccfg.create":"Create",
	}

)

func translationsjson() string {
	first := true

	body:="{"

	for k , v := range TRANSLATIONS {
		if first {
			first=false
		} else {
			body+=","
		}
		body+="\""+k+"\":\""+v+"\""
	}

	body+="}"

	return body
}