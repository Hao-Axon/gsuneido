// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

// +build !portable

package builtin

import (
	"unsafe"

	"github.com/apmckinlay/gsuneido/builtin/goc"
	"github.com/apmckinlay/gsuneido/builtin/heap"
	. "github.com/apmckinlay/gsuneido/runtime"
)

var comdlg32 = MustLoadDLL("comdlg32.dll")

// dll long ComDlg32:CommDlgExtendedError()
var commDlgExtendedError = comdlg32.MustFindProc("CommDlgExtendedError").Addr()
var _ = builtin0("CommDlgExtendedError()",
	func() Value {
		rtn := goc.Syscall0(commDlgExtendedError)
		return intRet(rtn)
	})

// dll bool ComDlg32:PrintDlg(PRINTDLG* printdlg)
var printDlg = comdlg32.MustFindProc("PrintDlgA").Addr()
var _ = builtin1("PrintDlg(printdlg)",
	func(a Value) Value {
		defer heap.FreeTo(heap.CurSize())
		p := heap.Alloc(nPRINTDLG)
		pd := (*PRINTDLG)(p)
		*pd = PRINTDLG{
			lStructSize:         uint32(nPRINTDLG),
			hwndOwner:           getUintptr(a, "hwndOwner"),
			hDevMode:            getUintptr(a, "hDevMode"),
			hDevNames:           getUintptr(a, "hDevNames"),
			hDC:                 getUintptr(a, "hDC"),
			Flags:               getInt32(a, "Flags"),
			nFromPage:           getInt16(a, "nFromPage"),
			nToPage:             getInt16(a, "nToPage"),
			nMinPage:            getInt16(a, "nMinPage"),
			nMaxPage:            getInt16(a, "nMaxPage"),
			nCopies:             getInt16(a, "nCopies"),
			hInstance:           getUintptr(a, "hInstance"),
			lCustData:           getUintptr(a, "lCustData"),
			lpPrintTemplateName: getStr(a, "lpPrintTemplateName"),
			lpSetupTemplateName: getStr(a, "lpSetupTemplateName"),
			hPrintTemplate:      getUintptr(a, "hPrintTemplate"),
			hSetupTemplate:      getUintptr(a, "hSetupTemplate"),
		}
		rtn := goc.Syscall1(printDlg,
			uintptr(p))
		a.Put(nil, SuStr("hwndOwner"), IntVal(int(pd.hwndOwner)))
		a.Put(nil, SuStr("hDevMode"), IntVal(int(pd.hDevMode)))
		a.Put(nil, SuStr("hDevNames"), IntVal(int(pd.hDevNames)))
		a.Put(nil, SuStr("hDC"), IntVal(int(pd.hDC)))
		a.Put(nil, SuStr("Flags"), IntVal(int(pd.Flags)))
		a.Put(nil, SuStr("nFromPage"), IntVal(int(pd.nFromPage)))
		a.Put(nil, SuStr("nToPage"), IntVal(int(pd.nToPage)))
		a.Put(nil, SuStr("nMinPage"), IntVal(int(pd.nMinPage)))
		a.Put(nil, SuStr("nMaxPage"), IntVal(int(pd.nMaxPage)))
		a.Put(nil, SuStr("nCopies"), IntVal(int(pd.nCopies)))
		a.Put(nil, SuStr("hInstance"), IntVal(int(pd.hInstance)))
		a.Put(nil, SuStr("lCustData"), IntVal(int(pd.lCustData)))
		a.Put(nil, SuStr("hPrintTemplate"), IntVal(int(pd.hPrintTemplate)))
		a.Put(nil, SuStr("hSetupTemplate"), IntVal(int(pd.hSetupTemplate)))
		return boolRet(rtn)
	})

type PRINTDLG struct {
	lStructSize         uint32
	hwndOwner           HANDLE
	hDevMode            HANDLE
	hDevNames           HANDLE
	hDC                 HANDLE
	Flags               int32
	nFromPage           int16
	nToPage             int16
	nMinPage            int16
	nMaxPage            int16
	nCopies             int16
	hInstance           HANDLE
	lCustData           HANDLE
	lpfnPrintHook       HANDLE
	lpfnSetupHook       HANDLE
	lpPrintTemplateName *byte
	lpSetupTemplateName *byte
	hPrintTemplate      HANDLE
	hSetupTemplate      HANDLE
}

const nPRINTDLG = unsafe.Sizeof(PRINTDLG{})

// dll bool ComDlg32:PageSetupDlg(PAGESETUPDLG* pagesetupdlg)
var pageSetupDlg = comdlg32.MustFindProc("PageSetupDlgA").Addr()
var _ = builtin1("PageSetupDlg(pagesetupdlg)",
	func(a Value) Value {
		defer heap.FreeTo(heap.CurSize())
		p := heap.Alloc(nPAGESETUPDLG)
		psd := (*PAGESETUPDLG)(p)
		*psd = PAGESETUPDLG{
			lStructSize:             uint32(nPAGESETUPDLG),
			ptPaperSize:             getPoint(a, "ptPaperSize"),
			rtMinMargin:             getRect(a, "rtMinMargin"),
			rtMargin:                getRect(a, "rtMargin"),
			hwndOwner:               getUintptr(a, "hwndOwner"),
			hDevMode:                getUintptr(a, "hDevMode"),
			hDevNames:               getUintptr(a, "hDevNames"),
			Flags:                   getInt32(a, "Flags"),
			hInstance:               getUintptr(a, "hInstance"),
			lCustData:               getUintptr(a, "lCustData"),
			lpfnPageSetupHook:       0,
			lpfnPagePaintHook:       0,
			lpPageSetupTemplateName: getStr(a, "lpPageSetupTemplateName"),
			hPageSetupTemplate:      getUintptr(a, "hPageSetupTemplate"),
		}
		rtn := goc.Syscall1(pageSetupDlg,
			uintptr(p))
		a.Put(nil, SuStr("hwndOwner"), IntVal(int(psd.hwndOwner)))
		a.Put(nil, SuStr("hDevMode"), IntVal(int(psd.hDevMode)))
		a.Put(nil, SuStr("hDevNames"), IntVal(int(psd.hDevNames)))
		a.Put(nil, SuStr("Flags"), IntVal(int(psd.Flags)))
		a.Put(nil, SuStr("ptPaperSize"),
			pointToOb(&psd.ptPaperSize, a.Get(nil, SuStr("ptPaperSize"))))
		a.Put(nil, SuStr("rtMinMargin"),
			rectToOb(&psd.rtMinMargin, a.Get(nil, SuStr("rtMinMargin"))))
		a.Put(nil, SuStr("rtMargin"),
			rectToOb(&psd.rtMargin, a.Get(nil, SuStr("rtMargin"))))
		a.Put(nil, SuStr("hInstance"), IntVal(int(psd.hInstance)))
		a.Put(nil, SuStr("lCustData"), IntVal(int(psd.lCustData)))
		a.Put(nil, SuStr("lpfnPageSetupHook"), IntVal(int(psd.lpfnPageSetupHook)))
		a.Put(nil, SuStr("lpfnPagePaintHook"), IntVal(int(psd.lpfnPagePaintHook)))
		a.Put(nil, SuStr("hPageSetupTemplate"), IntVal(int(psd.hPageSetupTemplate)))
		return boolRet(rtn)
	})

type PAGESETUPDLG struct {
	lStructSize             uint32
	hwndOwner               HANDLE
	hDevMode                HANDLE
	hDevNames               HANDLE
	Flags                   int32
	ptPaperSize             POINT
	rtMinMargin             RECT
	rtMargin                RECT
	hInstance               HANDLE
	lCustData               uintptr
	lpfnPageSetupHook       HANDLE
	lpfnPagePaintHook       HANDLE
	lpPageSetupTemplateName *byte
	hPageSetupTemplate      HANDLE
}

const nPAGESETUPDLG = unsafe.Sizeof(PAGESETUPDLG{})

// dll bool ComDlg32:GetSaveFileName(OPENFILENAME* ofn)
var getSaveFileName = comdlg32.MustFindProc("GetSaveFileNameA").Addr()
var _ = builtin1("GetSaveFileName(a)",
	func(a Value) Value {
		defer heap.FreeTo(heap.CurSize())
		p, buf, bufsize := buildOPENFILENAME(a)
		rtn := goc.Syscall1(getSaveFileName,
			uintptr(p))
		if rtn != 0 {
			a.Put(nil, SuStr("file"), SuStr(heap.GetStrZ(buf, bufsize)))
		}
		return boolRet(rtn)
	})

// dll bool ComDlg32:GetOpenFileName(OPENFILENAME* ofn)
var getOpenFileName = comdlg32.MustFindProc("GetOpenFileNameA").Addr()
var _ = builtin1("GetOpenFileName(a)",
	func(a Value) Value {
		defer heap.FreeTo(heap.CurSize())
		p, buf, bufsize := buildOPENFILENAME(a)
		rtn := goc.Syscall1(getOpenFileName,
			uintptr(p))
		if rtn != 0 {
			a.Put(nil, SuStr("file"), bufStrZ2(buf, uintptr(bufsize)))
		}
		return boolRet(rtn)
	})

func buildOPENFILENAME(a Value) (p unsafe.Pointer, buf unsafe.Pointer, bufsize int) {
	bufsize = getInt(a, "maxFile")
	file := ToStr(a.Get(nil, SuStr("file")))
	buf = heap.Copy(file, bufsize)
	p = heap.Alloc(nOPENFILENAME)
	*(*OPENFILENAME)(p) = OPENFILENAME{
		structSize: int32(nOPENFILENAME),
		hwndOwner:  getUintptr(a, "hwndOwner"),
		file:       (*byte)(buf),
		maxFile:    int32(bufsize),
		filter:     getStr(a, "filter"),
		flags:      getInt32(a, "flags"),
		defExt:     getStr(a, "defExt"),
		initialDir: getStr(a, "initialDir"),
		title:      getStr(a, "title"),
	}
	return
}

type OPENFILENAME struct {
	structSize     int32
	hwndOwner      HANDLE
	instance       HANDLE
	filter         *byte
	customFilter   *byte
	nMaxCustFilter int32
	nFilterIndex   int32
	file           *byte
	maxFile        int32
	fileTitle      *byte
	maxFileTitle   int32
	initialDir     *byte
	title          *byte
	flags          int32
	fileOffset     int16
	fileExtension  int16
	defExt         *byte
	custData       HANDLE
	hook           HANDLE
	templateName   *byte
	pvReserved     uintptr
	dwReserved     int32
	FlagsEx        int32
}

const nOPENFILENAME = unsafe.Sizeof(OPENFILENAME{})

// dll bool ComDlg32:ChooseColor(CHOOSECOLOR* x)
var chooseColor = comdlg32.MustFindProc("ChooseColorA").Addr()
var _ = builtin1("ChooseColor(x)",
	func(a Value) Value {
		defer heap.FreeTo(heap.CurSize())
		custColors := (*CustColors)(heap.Alloc(nCustColors * int32Size))
		ccs := a.Get(nil, SuStr("custColors"))
		for i := 0; i < nCustColors; i++ {
			if x := ccs.Get(nil, SuInt(i)); x != nil {
				custColors[i] = int32(ToInt(x))
			}
		}
		p := heap.Alloc(nCHOOSECOLOR)
		cc := (*CHOOSECOLOR)(p)
		*cc = CHOOSECOLOR{
			size:         int32(nCHOOSECOLOR),
			owner:        getUintptr(a, "owner"),
			instance:     getUintptr(a, "instance"),
			rgbResult:    getInt32(a, "rgbResult"),
			custColors:   custColors,
			flags:        getInt32(a, "flags"),
			custData:     getUintptr(a, "custData"),
			hook:         getUintptr(a, "hook"),
			templateName: getStr(a, "templateName"),
		}
		rtn := goc.Syscall1(chooseColor,
			uintptr(p))
		a.Put(nil, SuStr("rgbResult"), IntVal(int(cc.rgbResult)))
		a.Put(nil, SuStr("flags"), IntVal(int(cc.flags)))
		for i := 0; i < nCustColors; i++ {
			ccs.Put(nil, SuInt(i), IntVal(int(custColors[i])))
		}
		return boolRet(rtn)
	})

type CHOOSECOLOR struct {
	size         int32
	owner        HANDLE
	instance     HANDLE
	rgbResult    int32
	custColors   *CustColors
	flags        int32
	custData     HANDLE
	hook         HANDLE
	templateName *byte
}

const nCHOOSECOLOR = unsafe.Sizeof(CHOOSECOLOR{})

const nCustColors = 16

type CustColors [nCustColors]int32

// dll bool ComDlg32:ChooseFont(CHOOSEFONT* cf)
var chooseFont = comdlg32.MustFindProc("ChooseFontA").Addr()
var _ = builtin1("ChooseFont(cf)",
	func(a Value) Value {
		defer heap.FreeTo(heap.CurSize())
		lf := (*LOGFONT)(heap.Alloc(nLOGFONT))
		lfob := a.Get(nil, SuStr("lpLogFont"))
		*lf = LOGFONT{
			lfHeight:         getInt32(lfob, "lfHeight"),
			lfWidth:          getInt32(lfob, "lfWidth"),
			lfEscapement:     getInt32(lfob, "lfEscapement"),
			lfOrientation:    getInt32(lfob, "lfOrientation"),
			lfWeight:         getInt32(lfob, "lfWeight"),
			lfItalic:         byte(getInt(lfob, "lfItalic")),
			lfUnderline:      byte(getInt(lfob, "lfUnderline")),
			lfStrikeOut:      byte(getInt(lfob, "lfStrikeOut")),
			lfCharSet:        byte(getInt(lfob, "lfCharSet")),
			lfOutPrecision:   byte(getInt(lfob, "lfOutPrecision")),
			lfClipPrecision:  byte(getInt(lfob, "lfClipPrecision")),
			lfQuality:        byte(getInt(lfob, "lfQuality")),
			lfPitchAndFamily: byte(getInt(lfob, "lfPitchAndFamily")),
		}
		getStrZbs(lfob, "lfFaceName", lf.lfFaceName[:])
		p := heap.Alloc(nCHOOSEFONT)
		*(*CHOOSEFONT)(p) = CHOOSEFONT{
			lStructSize:    uint32(nCHOOSEFONT),
			hwndOwner:      getUintptr(a, "hwndOwner"),
			hDC:            getUintptr(a, "hDC"),
			lpLogFont:      lf,
			iPointSize:     getInt32(a, "iPointSize"),
			Flags:          getInt32(a, "Flags"),
			rgbColors:      getInt32(a, "rgbColors"),
			lCustData:      getUintptr(a, "lCustData"),
			lpfnHook:       getUintptr(a, "lpfnHook"),
			lpTemplateName: getStr(a, "lpTemplateName"),
			hInstance:      getUintptr(a, "hInstance"),
			lpszStyle:      getStr(a, "lpszStyle"),
			nFontType:      getInt16(a, "nFontType"),
			nSizeMin:       getInt32(a, "nSizeMin"),
			nSizeMax:       getInt32(a, "nSizeMax"),
		}
		rtn := goc.Syscall1(chooseFont,
			uintptr(p))
		lfob.Put(nil, SuStr("lfHeight"), IntVal(int(lf.lfHeight)))
		lfob.Put(nil, SuStr("lfWidth"), IntVal(int(lf.lfWidth)))
		lfob.Put(nil, SuStr("lfEscapement"), IntVal(int(lf.lfEscapement)))
		lfob.Put(nil, SuStr("lfOrientation"), IntVal(int(lf.lfOrientation)))
		lfob.Put(nil, SuStr("lfWeight"), IntVal(int(lf.lfWeight)))
		lfob.Put(nil, SuStr("lfItalic"), IntVal(int(lf.lfItalic)))
		lfob.Put(nil, SuStr("lfUnderline"), IntVal(int(lf.lfUnderline)))
		lfob.Put(nil, SuStr("lfStrikeOut"), IntVal(int(lf.lfStrikeOut)))
		lfob.Put(nil, SuStr("lfCharSet"), IntVal(int(lf.lfCharSet)))
		lfob.Put(nil, SuStr("lfOutPrecision"), IntVal(int(lf.lfOutPrecision)))
		lfob.Put(nil, SuStr("lfClipPrecision"), IntVal(int(lf.lfClipPrecision)))
		lfob.Put(nil, SuStr("lfQuality"), IntVal(int(lf.lfQuality)))
		lfob.Put(nil, SuStr("lfPitchAndFamily"), IntVal(int(lf.lfPitchAndFamily)))
		lfob.Put(nil, SuStr("lfPitchAndFamily"), IntVal(int(lf.lfPitchAndFamily)))
		lfob.Put(nil, SuStr("lfFaceName"), bsStrZ(lf.lfFaceName[:]))
		return boolRet(rtn)
	})

type CHOOSEFONT struct {
	lStructSize    uint32
	hwndOwner      HANDLE
	hDC            HANDLE
	lpLogFont      *LOGFONT
	iPointSize     int32
	Flags          int32
	rgbColors      int32
	lCustData      uintptr
	lpfnHook       HANDLE
	lpTemplateName *byte
	hInstance      HANDLE
	lpszStyle      *byte
	nFontType      int16
	_              int16 // padding
	nSizeMin       int32
	nSizeMax       int32
	_              int32 // padding
}

const nCHOOSEFONT = unsafe.Sizeof(CHOOSEFONT{})

// dll HRESULT ComDlg32:PrintDlgEx(PRINTDLGEX* printdlgex)
var printDlgEx = comdlg32.MustFindProc("PrintDlgExA").Addr()
var _ = builtin1("PrintDlgEx(printdlgex)",
	func(a Value) Value {
		defer heap.FreeTo(heap.CurSize())
		p := heap.Alloc(nPRINTDLGEX)
		pd := (*PRINTDLGEX)(p)
		*pd = PRINTDLGEX{
			lStructSize:         int32(nPRINTDLGEX),
			hwndOwner:           getUintptr(a, "hwndOwner"),
			hDevMode:            getUintptr(a, "hDevMode"),
			hDevNames:           getUintptr(a, "hDevNames"),
			hDC:                 getUintptr(a, "hDC"),
			Flags:               getInt32(a, "Flags"),
			Flags2:              getInt32(a, "Flags2"),
			ExclusionFlags:      getInt32(a, "ExclusionFlags"),
			nMinPage:            getInt32(a, "nMinPage"),
			nMaxPage:            getInt32(a, "nMaxPage"),
			nCopies:             getInt32(a, "nCopies"),
			hInstance:           getUintptr(a, "hInstance"),
			lpPrintTemplateName: getStr(a, "lpPrintTemplateName"),
			nStartPage:          getInt32(a, "nStartPage"),
			dwResultAction:      getInt32(a, "dwResultAction"),
		}
		prob := a.Get(nil, SuStr("lpPageRanges"))
		var pr *PRINTPAGERANGE
		if prob != nil {
			pr = (*PRINTPAGERANGE)(heap.Alloc(nPRINTPAGERANGE))
			*pr = PRINTPAGERANGE{
				nFromPage: getInt32(prob, "nFromPage"),
				nToPage:   getInt32(prob, "nToPage"),
			}
			pd.lpPageRanges = pr
			pd.nPageRanges = 1
			pd.nMaxPageRanges = 1
		}
		rtn := goc.Syscall1(printDlgEx,
			uintptr(p))
		a.Put(nil, SuStr("hwndOwner"), IntVal(int(pd.hwndOwner)))
		a.Put(nil, SuStr("hDevMode"), IntVal(int(pd.hDevMode)))
		a.Put(nil, SuStr("hDevNames"), IntVal(int(pd.hDevNames)))
		a.Put(nil, SuStr("hDC"), IntVal(int(pd.hDC)))
		a.Put(nil, SuStr("Flags"), IntVal(int(pd.Flags)))
		a.Put(nil, SuStr("Flags2"), IntVal(int(pd.Flags2)))
		a.Put(nil, SuStr("ExclusionFlags"), IntVal(int(pd.ExclusionFlags)))
		a.Put(nil, SuStr("nPageRanges"), IntVal(int(pd.nPageRanges)))
		a.Put(nil, SuStr("nMaxPageRanges"), IntVal(int(pd.nMaxPageRanges)))
		if prob != nil {
			prob.Put(nil, SuStr("nFromPage"), IntVal(int(pr.nFromPage)))
			prob.Put(nil, SuStr("nToPage"), IntVal(int(pr.nToPage)))
		}
		a.Put(nil, SuStr("nMinPage"), IntVal(int(pd.nMinPage)))
		a.Put(nil, SuStr("nMaxPage"), IntVal(int(pd.nMaxPage)))
		a.Put(nil, SuStr("nCopies"), IntVal(int(pd.nCopies)))
		a.Put(nil, SuStr("hInstance"), IntVal(int(pd.hInstance)))
		a.Put(nil, SuStr("nStartPage"), IntVal(int(pd.nStartPage)))
		a.Put(nil, SuStr("dwResultAction"), IntVal(int(pd.dwResultAction)))
		return intRet(rtn)
	})

type PRINTDLGEX struct {
	lStructSize         int32
	hwndOwner           HANDLE
	hDevMode            HANDLE
	hDevNames           HANDLE
	hDC                 HANDLE
	Flags               int32
	Flags2              int32
	ExclusionFlags      int32
	nPageRanges         int32
	nMaxPageRanges      int32
	lpPageRanges        *PRINTPAGERANGE
	nMinPage            int32
	nMaxPage            int32
	nCopies             int32
	hInstance           HANDLE
	lpPrintTemplateName *byte
	lpCallback          uintptr
	nPropertyPages      int32
	lphPropertyPages    uintptr
	nStartPage          int32
	dwResultAction      int32
}

const nPRINTDLGEX = unsafe.Sizeof(PRINTDLGEX{})

type PRINTPAGERANGE struct {
	nFromPage int32
	nToPage   int32
}

const nPRINTPAGERANGE = unsafe.Sizeof(PRINTPAGERANGE{})
