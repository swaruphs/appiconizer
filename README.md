# Appiconizer

Resize app icons for iOS and Android from command line. Icons are stored under
folder named 'appiconizer <current date>'

## Install

You can install Appiconizer using the [Homebrew](https://github.com/Homebrew/homebrew/) package manager on Mac OS X:
```shell
$ brew tap swaruphs/tap
$ brew install appiconizer

```

## Usage manual
```console
Usage: appiconizer <command> [command flags]

create command:
  -source string
      Source image path. Preferred size is 1024X1024
  -target string
      Target location to store the generated files.
  -device string
      Specify device (ios/android/all). Defaults to all
  -zip bool
      want to generate as a zip folder.

examples:
  appiconizer create -source /Users/user/Desktop/sample.png
```

#### `-source`
Specifies the source image file. Preferred size is 1024X1024.

#### `-target`
Specifies the target folder path. Optional.
Defaults to storing in the same path as image file.

#### `-device`
Specifies list of devices for generating icons. Option can be one from
ios/android/all. Defaults to all.

#### `-zip`
Optional boolean flag to generate icons as zip folder.

##License 
MIT License. Refer to LICENSE.md
