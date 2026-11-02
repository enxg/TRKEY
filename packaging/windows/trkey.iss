; TRKEY Windows installer (Inno Setup 6.3+).
;
;   iscc /DAppVersion=1.0.0 /DBuildDir=build trkey.iss
;
; BuildDir must contain amd64\trkey.dll and wow64\trkey.dll (kbdutool -m / -o).

#ifndef AppVersion
  #define AppVersion "0.0.0-dev"
#endif

#ifndef BuildDir
  #define BuildDir "build"
#endif

#define LayoutName "TRKEY (US)"

[Setup]
; AppId must never change, it ties upgrades and uninstall together
AppId={{C027357E-1348-4BD8-A11F-62C1A0E46A0D}
AppName=TRKEY
AppVersion={#AppVersion}
AppVerName=TRKEY {#AppVersion}
AppPublisher=Enes Genç
AppPublisherURL=https://github.com/enxg/TRKEY
AppSupportURL=https://github.com/enxg/TRKEY/issues
UninstallDisplayName={#LayoutName} keyboard layout
DefaultDirName={autopf}\TRKEY
DisableDirPage=yes
DisableProgramGroupPage=yes
DisableReadyPage=yes
PrivilegesRequired=admin
; x64 Windows only; ARM64 needs a native layout DLL that kbdutool can't build
ArchitecturesAllowed=x64os
ArchitecturesInstallIn64BitMode=x64os
LicenseFile=..\..\LICENSE
OutputBaseFilename=TRKEY-setup
WizardStyle=modern
Compression=lzma2
SolidCompression=yes
; keyboard DLLs are loaded by every app; never ask to close them
CloseApplications=no

[Tasks]
Name: english; Description: "Also add {#LayoutName} under English (United States)"; Flags: unchecked

[Files]
Source: "{#BuildDir}\amd64\trkey.dll"; DestDir: "{sys}"; Flags: ignoreversion uninsrestartdelete
Source: "{#BuildDir}\wow64\trkey.dll"; DestDir: "{syswow64}"; Flags: ignoreversion uninsrestartdelete
Source: "languages.ps1"; DestDir: "{app}"; Flags: ignoreversion

[Run]
Filename: "powershell.exe"; Parameters: "-NoProfile -ExecutionPolicy Bypass -File ""{app}\languages.ps1"""; \
  StatusMsg: "Adding {#LayoutName} to your languages..."; Tasks: not english; \
  Flags: runhidden runasoriginaluser waituntilterminated
Filename: "powershell.exe"; Parameters: "-NoProfile -ExecutionPolicy Bypass -File ""{app}\languages.ps1"" -English"; \
  StatusMsg: "Adding {#LayoutName} to your languages..."; Tasks: english; \
  Flags: runhidden runasoriginaluser waituntilterminated

[UninstallRun]
; runs before the registry entry is removed, so the script can still find the layout id
Filename: "powershell.exe"; Parameters: "-NoProfile -ExecutionPolicy Bypass -File ""{app}\languages.ps1"" -Remove"; \
  RunOnceId: "RemoveLanguages"; Flags: runhidden waituntilterminated

[Messages]
FinishedLabel=Setup has installed {#LayoutName}. It's in your keyboard list under Turkish; switch with Win+Space.%n%nIf it doesn't show up yet, sign out and back in.

[Code]
const
  LayoutsKey = 'SYSTEM\CurrentControlSet\Control\Keyboard Layouts';
  LayoutFile = 'trkey.dll';

{ Returns the registry key name of the layout that uses trkey.dll, or '' if none. }
function FindKlid(): String;
var
  Names: TArrayOfString;
  I: Integer;
  F: String;
begin
  Result := '';
  if not RegGetSubkeyNames(HKLM64, LayoutsKey, Names) then
    Exit;
  for I := 0 to GetArrayLength(Names) - 1 do
    if RegQueryStringValue(HKLM64, LayoutsKey + '\' + Names[I], 'Layout File', F) and
       (CompareText(F, LayoutFile) = 0) then
    begin
      Result := Names[I];
      Exit;
    end;
end;

{ First free custom id for Turkish: a000041f, a001041f, ... }
function NewKlid(): String;
var
  I: Integer;
begin
  for I := 0 to $FFF do
  begin
    Result := Lowercase(Format('a%.3x041f', [I]));
    if not RegKeyExists(HKLM64, LayoutsKey + '\' + Result) then
      Exit;
  end;
  RaiseException('No free keyboard layout id');
end;

{ One higher than any existing "Layout Id"; Windows needs these to be unique. }
function NewLayoutId(): String;
var
  Names: TArrayOfString;
  I, V, Max: Integer;
  S: String;
begin
  Max := 0;
  if RegGetSubkeyNames(HKLM64, LayoutsKey, Names) then
    for I := 0 to GetArrayLength(Names) - 1 do
      if RegQueryStringValue(HKLM64, LayoutsKey + '\' + Names[I], 'Layout Id', S) then
      begin
        V := StrToIntDef('$' + S, 0);
        if V > Max then
          Max := V;
      end;
  Result := Lowercase(Format('%.4x', [Max + 1]));
end;

{ A loaded DLL can't be overwritten but can be renamed. Move the old one aside
  so the new one is used right away; the old file goes on the next reboot. }
procedure MoveAside(const Path: String);
var
  Old: String;
  I: Integer;
begin
  if not FileExists(Path) then
    Exit;

  I := 0;
  repeat
    Old := Path + '.old' + IntToStr(I);
    I := I + 1;
  until (not FileExists(Old)) or DeleteFile(Old);

  if RenameFile(Path, Old) then
    if not DeleteFile(Old) then
      RestartReplace(Old, '');
end;

procedure RegisterLayout();
var
  Klid, Key: String;
begin
  Klid := FindKlid();
  if Klid = '' then
  begin
    Klid := NewKlid();
    Key := LayoutsKey + '\' + Klid;
    RegWriteStringValue(HKLM64, Key, 'Layout Id', NewLayoutId());
  end;

  Key := LayoutsKey + '\' + Klid;
  RegWriteStringValue(HKLM64, Key, 'Layout File', LayoutFile);
  RegWriteStringValue(HKLM64, Key, 'Layout Text', '{#LayoutName}');
  Log('Registered keyboard layout ' + Klid);
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  Redir: Boolean;
begin
  if CurStep = ssInstall then
  begin
    Redir := EnableFsRedirection(False);
    try
      MoveAside(ExpandConstant('{sys}\' + LayoutFile));
      MoveAside(ExpandConstant('{syswow64}\' + LayoutFile));
    finally
      EnableFsRedirection(Redir);
    end;
  end;

  if CurStep = ssPostInstall then
    RegisterLayout();
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  Klid: String;
begin
  if CurUninstallStep = usPostUninstall then
  begin
    Klid := FindKlid();
    if Klid <> '' then
      RegDeleteKeyIncludingSubkeys(HKLM64, LayoutsKey + '\' + Klid);
  end;
end;
