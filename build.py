import os
import json
import shutil

os.makedirs('build', exist_ok=True)

with open('build.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

build = data["BuildType"]

if build == "DEV":
    os.makedirs('build/Debug', exist_ok=True)
    os.system('go build -x -v -o build/Debug/himera.exe && build\\Debug\\himera.exe')

if build == "DEBUG":
    os.makedirs('build/Debug', exist_ok=True)

    curpath = os.getcwd()

    os.system('go build -x -v -o build/Debug/himera.exe')

    shutil.copytree(os.path.join(curpath, 'HGD\\ttf'), os.path.join(curpath, 'build\\Debug\\HGD\\ttf'))

if build == "RELEASE":
    os.makedirs('build/Release', exist_ok=True)
    
    curpath = os.getcwd()

    # Windows
    os.makedirs('build/Release/Windows', exist_ok=True)
    os.environ['GOOS'] = 'windows'
    os.environ['GOARCH'] = 'amd64'
    os.system('go build -ldflags="-s -w -H windowsgui" -x -v -o build/Release/Windows/himera_x64_86.exe')

    shutil.copytree(os.path.join(curpath, 'HGD\\ttf'), os.path.join(curpath, 'build\\Release\\Windows\\HGD\\ttf'))

    # WiP Linux
    # os.makedirs('build/Release/Linux', exist_ok=True)
    # CC=x86_64-linux-gnu-gcc CGO_ENABLED=1
    # os.environ['GOOS'] = 'linux'
    # os.environ['GOARCH'] = 'amd64'
    # os.system('go build -ldflags="-s -w" -x -v -o build/Release/Linux/himera_x64_86')
    # Wip MacOS ...
    # Wip Android ....

