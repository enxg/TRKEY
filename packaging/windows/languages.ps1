# Adds TRKEY to (or removes it from) the current user's Windows language list.
# Run by the installer as the user who started setup, not as the elevated admin.
#
#   languages.ps1             add under Turkish
#   languages.ps1 -English    add under Turkish and English (United States)
#   languages.ps1 -Remove     remove from every language

param(
    [switch]$English,
    [switch]$Remove
)

$ErrorActionPreference = 'Stop'

$layouts = 'HKLM:\SYSTEM\CurrentControlSet\Control\Keyboard Layouts'
$klid = Get-ChildItem $layouts |
    Where-Object { (Get-ItemProperty $_.PSPath).'Layout File' -eq 'trkey.dll' } |
    Select-Object -First 1 -ExpandProperty PSChildName

if (-not $klid) {
    throw 'TRKEY is not registered (no keyboard layout uses trkey.dll)'
}

$klid = $klid.ToUpper()
$list = Get-WinUserLanguageList

if ($Remove) {
    foreach ($lang in $list) {
        $tips = @($lang.InputMethodTips | Where-Object { $_ -like "*:$klid" })
        foreach ($tip in $tips) {
            [void]$lang.InputMethodTips.Remove($tip)
        }
    }

    Set-WinUserLanguageList $list -Force
    return
}

function Add-Layout([string[]]$tags, [string]$langId) {
    $lang = $list | Where-Object { $tags -contains $_.LanguageTag } | Select-Object -First 1

    if (-not $lang) {
        $list.Add($tags[0])
        $lang = $list | Where-Object { $tags -contains $_.LanguageTag } | Select-Object -First 1
        # a newly added language comes with its default keyboard; TRKEY replaces it
        $lang.InputMethodTips.Clear()
    }

    $tip = "${langId}:$klid"
    if ($lang.InputMethodTips -notcontains $tip) {
        $lang.InputMethodTips.Add($tip)
    }
}

# Windows reports Turkish as "tr" once added
Add-Layout @('tr-TR', 'tr') '041F'

if ($English) {
    Add-Layout @('en-US') '0409'
}

Set-WinUserLanguageList $list -Force
