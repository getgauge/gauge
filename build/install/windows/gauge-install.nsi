; This script builds gauge installation package
; It needs the following arguments to be set before calling compile
;    PRODUCT_VERSION -> Version of Gauge
;    GAUGE_DISTRIBUTABLES_DIR -> Directory where gauge distributables are available. It should not end with \
;    OUTPUT_FILE_NAME -> Name of the setup file

; HM NIS Edit Wizard helper defines
!define PRODUCT_NAME "Gauge"
!define PRODUCT_PUBLISHER "Gauge Community"
!define PRODUCT_WEB_SITE "https://gauge.org"
!define PRODUCT_DIR_REGKEY "Software\Microsoft\Windows\CurrentVersion\App Paths\gauge.exe"
!define PRODUCT_UNINST_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}"
!define PRODUCT_UNINST_ROOT_KEY "HKLM"
!define MUI_FINISHPAGE_LINK "Click here to read the Gauge Reference Documentation"
!define MUI_FINISHPAGE_LINK_LOCATION "https://docs.gauge.org"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_COMPONENTSPAGE_TEXT_COMPLIST "Additional plugins can be installed using the command 'gauge install <plugin>'"
; MUI 1.67 compatible ------
!include "MUI.nsh"
!include "MUI2.nsh"
!include "x64.nsh"
!include "winmessages.nsh"
!include "FileFunc.nsh" ;For GetOptions
!include "WordFunc.nsh"
!include "nsDialogs.nsh"

!define Explode "!insertmacro Explode"

!macro  Explode Length  Separator   String
  Push    `${Separator}`
  Push    `${String}`
  Call    Explode
  Pop     `${Length}`
!macroend

; MUI Settings
!define MUI_ABORTWARNING
!define MUI_ICON "gauge.ico"
!define MUI_UNICON "gauge.ico"
!define env_hklm 'HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"'
; Welcome page
!insertmacro MUI_PAGE_WELCOME
; License page
!insertmacro MUI_PAGE_LICENSE "license.txt"
; Plugin options page
!insertmacro MUI_PAGE_COMPONENTS
; Directory page
!insertmacro MUI_PAGE_DIRECTORY
; Instfiles page
!insertmacro MUI_PAGE_INSTFILES
; Finish page
!insertmacro MUI_PAGE_FINISH

; Uninstaller pages
!insertmacro MUI_UNPAGE_INSTFILES

; Language files
!insertmacro MUI_LANGUAGE "English"

; MUI end ------

SpaceTexts none

BrandingText "${PRODUCT_NAME} ${PRODUCT_VERSION}  |  ${PRODUCT_PUBLISHER}"

Name "${PRODUCT_NAME} ${PRODUCT_VERSION}"
OutFile "${OUTPUT_FILE_NAME}"
ShowUnInstDetails show

Section "Gauge" SEC_GAUGE
  SectionIn RO
  SetOutPath "$INSTDIR\bin"
  SetOverwrite on
  File /r "${GAUGE_DISTRIBUTABLES_DIR}\*"
SectionEnd

SectionGroup /e "Language Plugins" SEC_LANGUAGES
  Section /o "Java" SEC_JAVA
  SectionEnd
  Section /o "Ruby" SEC_RUBY
  SectionEnd
  Section /o "JavaScript" SEC_JAVASCRIPT
  SectionEnd
  Section /o "Python" SEC_PYTHON
  SectionEnd
  Section /o "Dotnet" SEC_DOTNET
  SectionEnd
SectionGroupEnd

SectionGroup /e "Reporting Plugins" SEC_REPORTS
  Section "HTML" SEC_HTML
    SectionIn RO
  SectionEnd
  Section /o "XML" SEC_XML
  SectionEnd
  Section /o "Spectacle" SEC_SPECTACLE
  SectionEnd
SectionGroupEnd


!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_GAUGE} "Will install Gauge Core (gauge.exe)."
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_LANGUAGES} "Check to install language runners that needs to be installed. You need at least one language runner to run Gauge specs."
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_JAVA} "Java language runner, enables writing implementations using Java."
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_RUBY} "Ruby language runner, enables writing implementations using Ruby."
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_JAVASCRIPT} "JavaScript language runner, enables writing implementations using JavaScript."
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_PYTHON} "Python language runner, enables writing implementations using Python."
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_DOTNET} "Dotnet core language runner, enables writing implementations using Dotnet core."
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_REPORTS} "Check to install reporting plugins. HTML report plugin is installed by default."
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_HTML} "Generates HTML report of Gauge spec run."
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_XML} "Generates JUnit style XML report of Gauge spec run."
  !insertmacro MUI_DESCRIPTION_TEXT ${SEC_SPECTACLE} "Generates static HTML from Spec files, allows filtering/navigation of Specifications."
!insertmacro MUI_FUNCTION_DESCRIPTION_END

function .onInit
  ${If} $INSTDIR == ""
    ${If} ${RunningX64}
      SetRegView 64
      StrCpy $INSTDIR "$PROGRAMFILES64\Gauge"
    ${Else}
      StrCpy $INSTDIR "$PROGRAMFILES\Gauge"
    ${EndIf}
  ${EndIf}
functionEnd

Section -AdditionalIcons
  SetOutPath $INSTDIR
  CreateDirectory "$SMPROGRAMS\Gauge"
  CreateShortCut "$SMPROGRAMS\Gauge\Uninstall.lnk" "$INSTDIR\uninst.exe"
SectionEnd

Section -Post
  File "update_path.ps1"
  WriteUninstaller "$INSTDIR\uninst.exe"
  WriteRegStr HKLM "${PRODUCT_DIR_REGKEY}" "" "$INSTDIR\bin\gauge.exe"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayName" "$(^Name)"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "UninstallString" "$INSTDIR\uninst.exe"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayIcon" "$INSTDIR\bin\gauge.exe"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayVersion" "${PRODUCT_VERSION}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "URLInfoAbout" "${PRODUCT_WEB_SITE}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "Publisher" "${PRODUCT_PUBLISHER}"
  nsExec::ExecToLog 'powershell -ExecutionPolicy Bypass -WindowStyle Hidden -File "$INSTDIR\update_path.ps1" -Add -Path "$INSTDIR\bin"'
  SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000"

  Dialer::GetConnectedState
  Pop $R0

  ${If} $R0 == 'online'
    SectionGetFlags ${SEC_JAVA} $R1
    SectionGetFlags ${SEC_RUBY} $R2
    SectionGetFlags ${SEC_JAVASCRIPT} $R3
    SectionGetFlags ${SEC_PYTHON} $R4
    SectionGetFlags ${SEC_DOTNET} $R5
    SectionGetFlags ${SEC_XML} $R6
    SectionGetFlags ${SEC_SPECTACLE} $R7

    ${If} $R1 == 1
      DetailPrint "Installing plugin : java"
      nsExec::ExecToLog 'gauge install java'
    ${EndIf}

    ${If} $R2 == 1
      DetailPrint "Installing plugin : ruby"
      nsExec::ExecToLog 'gauge install ruby'
    ${EndIf}

    ${If} $R3 == 1
      DetailPrint "Installing plugin : javascript"
      nsExec::ExecToLog 'gauge install js'
    ${EndIf}

    ${If} $R4 == 1
      DetailPrint "Installing plugin : python"
      nsExec::ExecToLog 'gauge install python'
    ${EndIf}

    ${If} $R5 == 1
      DetailPrint "Installing plugin : dotnet"
      nsExec::ExecToLog 'gauge install dotnet'
    ${EndIf}

    ${If} $R6 == 1
      DetailPrint "Installing plugin : xml-report"
      nsExec::ExecToLog 'gauge install xml-report'
    ${EndIf}

    ${If} $R7 == 1
      DetailPrint "Installing plugin : spectacle"
      nsExec::ExecToLog 'gauge install spectacle'
    ${EndIf}

    DetailPrint "Installing plugin : html-report"
    nsExec::ExecToLog 'gauge install html-report'
  ${Else}
    DetailPrint "[WARNING] Internet connection unavailable. Skipping plugins installation"
  ${EndIf}
SectionEnd

Function un.onUninstSuccess
  IfSilent +3 0
    HideWindow
      MessageBox MB_ICONINFORMATION|MB_OK "$(^Name) was successfully removed from your computer."
FunctionEnd

Function un.onInit
  IfSilent +3 0
    MessageBox MB_ICONQUESTION|MB_YESNO|MB_DEFBUTTON2 "Are you sure you want to completely remove $(^Name) and all of its components?" IDYES +2
  Abort
FunctionEnd

Section Uninstall
  Delete "$INSTDIR\uninst.exe"
  Delete "$INSTDIR\plugin-install.bat"
  Delete "$INSTDIR\update_path.ps1"
  RMDir /r "$INSTDIR\bin"
  Delete "$SMPROGRAMS\Gauge\Uninstall.lnk"
  DeleteRegKey ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}"
  DeleteRegKey HKLM "${PRODUCT_DIR_REGKEY}"
  DeleteRegValue ${env_hklm} GAUGE_ROOT
  RMDir "$INSTDIR"
  SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
  SetAutoClose true
SectionEnd

; Ref: http://nsis.sourceforge.net/Explode
Function Explode
  ; Initialize variables
  Var /GLOBAL explString
  Var /GLOBAL explSeparator
  Var /GLOBAL explStrLen
  Var /GLOBAL explSepLen
  Var /GLOBAL explOffset
  Var /GLOBAL explTmp
  Var /GLOBAL explTmp2
  Var /GLOBAL explTmp3
  Var /GLOBAL explArrCount

  ; Get input from user
  Pop $explString
  Pop $explSeparator

  ; Calculates initial values
  StrLen $explStrLen $explString
  StrLen $explSepLen $explSeparator
  StrCpy $explArrCount 1

  ${If}   $explStrLen <= 1          ;   If we got a single character
  ${OrIf} $explSepLen > $explStrLen ;   or separator is larger than the string,
    Push    $explString             ;   then we return initial string with no change
    Push    1                       ;   and set array's length to 1
    Return
  ${EndIf}

  ; Set offset to the last symbol of the string
  StrCpy $explOffset $explStrLen
  IntOp  $explOffset $explOffset - 1

  ; Clear temp string to exclude the possibility of appearance of occasional data
  StrCpy $explTmp   ""
  StrCpy $explTmp2  ""
  StrCpy $explTmp3  ""

  ; Loop until the offset becomes negative
  ${Do}
    ;   If offset becomes negative, it is time to leave the function
    ${IfThen} $explOffset == -1 ${|} ${ExitDo} ${|}

    ;   Remove everything before and after the searched part ("TempStr")
    StrCpy $explTmp $explString $explSepLen $explOffset

    ${If} $explTmp == $explSeparator
        ;   Calculating offset to start copy from
        IntOp   $explTmp2 $explOffset + $explSepLen ;   Offset equals to the current offset plus length of separator
        StrCpy  $explTmp3 $explString "" $explTmp2

        Push    $explTmp3                           ;   Throwing array item to the stack
        IntOp   $explArrCount $explArrCount + 1     ;   Increasing array's counter

        StrCpy  $explString $explString $explOffset 0   ;   Cutting all characters beginning with the separator entry
        StrLen  $explStrLen $explString
    ${EndIf}

    ${If} $explOffset = 0                       ;   If the beginning of the line met and there is no separator,
                                                ;   copying the rest of the string
        ${If} $explSeparator == ""              ;   Fix for the empty separator
            IntOp   $explArrCount   $explArrCount - 1
        ${Else}
            Push    $explString
        ${EndIf}
    ${EndIf}

    IntOp   $explOffset $explOffset - 1
  ${Loop}

  Push $explArrCount
"FunctionEnd"
