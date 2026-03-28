import Quickshell
import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

PanelWindow {
    id: root
    
    anchors {
        bottom: true
        left: true
        right: true
    }
    
    implicitHeight: 400
    
    exclusionMode: ExclusionMode.Ignore
    
    aboveWindows: true
    
    focusable: true

    color: "transparent"

    ListModel {
        id: wallpaperModel
    }

    Shortcut {
        sequence: "Escape"
        onActivated: Qt.quit()
    }

    function fetchWallpapers(queryParams) {
        var request = new XMLHttpRequest()
        request.open('GET', 'http://localhost:8080/api/search?' + queryParams, true)
        
        request.onreadystatechange = function() {
            if (request.readyState === XMLHttpRequest.DONE && request.status === 200) {
                var response = JSON.parse(request.responseText)
                wallpaperModel.clear() 
                
                for (var i = 0; i < response.data.length; i++) {
                    var item = response.data[i]
                    wallpaperModel.append({
                        "imageId": item.id,
                        "thumbUrl": item.thumbs.large,
                        "fullUrl": item.path
                    })
                }

                gallery.positionViewAtBeginning()
            }
        }
        request.send()
    }

    function setWallpaper(id, url) {
        var request = new XMLHttpRequest()
        var encodedUrl = encodeURIComponent(url)
        request.open('GET', 'http://localhost:8080/api/set?id=' + id + '&url=' + encodedUrl, true)
        request.send()
    }

    Component.onCompleted: {
        fetchWallpapers("linux") 
    }

    ColumnLayout {
        anchors.fill: parent
        anchors.margins: 20
        spacing: 20

        Rectangle {
            Layout.alignment: Qt.AlignHCenter
            Layout.preferredWidth: 600
            Layout.preferredHeight: 50
            color: "#282a36"
            radius: 25

            RowLayout {
                anchors.fill: parent
                anchors.margins: 10
                spacing: 15

                TextField {
                    id: searchInput
                    Layout.fillWidth: true
                    placeholderText: "Search wallpapers..."
                    color: "white"
                    background: Item {}
                    
                    onAccepted: {
                        if (text.trim() !== "") {
                            fetchWallpapers("q=" + encodeURIComponent(text))
                        }
                    }
                }

                Rectangle {
                    Layout.preferredWidth: 1
                    Layout.fillHeight: true
                    color: "#44475a"
                }

                Repeater {
                    model: ["#ea4c88", "#663399", "#333399", "#0099cc", "#66cccc", "#77cc33", "#ff9900", "#424153"]
                    
                    Rectangle {
                        Layout.preferredWidth: 30
                        Layout.preferredHeight: 30
                        radius: 15 
                        color: modelData
                        border.color: "#ffffff"
                        border.width: 1
                        
                        MouseArea {
                            anchors.fill: parent
                            cursorShape: Qt.PointingHandCursor
                            onClicked: {
                                var hexCode = modelData.replace("#", "")
                                fetchWallpapers("colors=" + hexCode)
                            }
                        }
                    }
                }
            }
        }

        ListView {
            id: gallery
            Layout.fillWidth: true
            Layout.fillHeight: true
            orientation: ListView.Horizontal
            spacing: 20
            model: wallpaperModel

            delegate: Rectangle {
                width: 400
                height: 250
                color: "black"
                radius: 12
                clip: true 

                Image {
                    anchors.fill: parent
                    source: thumbUrl
                    fillMode: Image.PreserveAspectCrop
                    asynchronous: true 

                    MouseArea {
                        anchors.fill: parent
                        cursorShape: Qt.PointingHandCursor
                        onClicked: {
                            console.log("Applying wallpaper ID: " + imageId)
                            root.setWallpaper(imageId, fullUrl)
                        }
                    }
                }
            }
        }
    }
}