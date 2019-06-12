package util_test

import (
	"testing"

	"0chain/gosdk/util"
)

func TestGetFileConfig(t *testing.T) {
	fileStr :=
		`{
        "Name": "sample.txt",
        "Path": "/",
        "Size": 200,
        "Type": "gzip"
        }`
	file, err := util.GetFileConfig(fileStr)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if file.Name != "sample.txt" || file.Path != "/" || file.Size != 200 || file.Type != "gzip" {
		t.Fatalf("File param doesn't match")
	}
}

func TestCalculateDirHash(t *testing.T) {
	dirStr := `
    {
    "type" : "d",
    "name" : "/",
    "hash" : "",
    "children" : [
        {
            "type" : "d",
            "name" : "Folder1",
            "hash" : "",
            "children" : [
                {
                    "type" : "f",
                    "name" : "file1.jpg",
                    "hash" : "2445ca278e6814d1089ec874e9be4986a46e9ba08851ea096ce00ea4802928fa"
                },
                {
                    "type" : "f",
                    "name" : "file2.jpg",
                    "hash" : "9f94b4fda348a7572350d862661d4f7254524b6220d2950ccaf832c9c9a46812"
                }
            ]
        },
        {
            "type" : "d",
            "name" : "Folder2",
            "hash" : "",
            "children" : [
                {
                    "type" : "f",
                    "name" : "file3.jpg",
                    "hash" : "c713fcf78ce2d31d3f2356f26ab1acfcff09873d5ac2549677f202d9749407db"
                }
            ]
        }
    ]
    }
    `
	r, err := util.GetDirTreeFromJson(dirStr)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	_ = util.CalculateDirHash(&r)
	if r.Children[0].Name != "Folder1" || r.Children[0].Hash != "85e17e081758d1a439b37a53511d9ee4e8ee659013281b27f36c09f43f5f85df" {
		t.Fatalf("Folder1 dir hash wrong")
	}
	if r.Children[1].Name != "Folder2" || r.Children[1].Hash != "be1c868a00884ff6664a86b546409ed61ad82c0f23cf41431133238ed04415e4" {
		t.Fatalf("Folder2 dir hash wrong")
	}
	if r.Name != "/" || r.Hash != "78e99381aca6a5112ccbafee710f0aba7f05ceb43f9867797fc3df44a5572e98" {
		t.Fatalf("/ hash is wrong")
	}
	st := util.GetJsonFromDirTree(&r)
	if st == "{}" {
		t.Fatalf("%s", err.Error())
	}
}

func TestInsertFile(t *testing.T) {
	dirStr := `
    {
    "type" : "d",
    "name" : "/",
    "hash" : ""
    }`
	root, err := util.GetDirTreeFromJson(dirStr)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	_, err = util.InsertFile(&root, "/photo1.jpg", "photo1.jpghash", 10)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if len(root.Children) != 1 || root.Children[0].Name != "photo1.jpg" || root.Children[0].Hash != "photo1.jpghash" || root.Children[0].Size != 10 {
		t.Fatalf("Adding /photo1.jpg failed")
	}

	_, err = util.InsertFile(&root, "/photo2.jpg", "photo2.jpghash", 100)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if len(root.Children) != 2 || root.Children[1].Name != "photo2.jpg" || root.Children[1].Hash != "photo2.jpghash" {
		t.Fatalf("Adding /photo2.jpg failed")
	}

	_, err = util.InsertFile(&root, "/photo2.jpg", "photo2.jpghash", 999)
	if err == nil {
		t.Fatalf("%s", err.Error())
	}

	_, err = util.InsertFile(&root, "/photo1/photo2.jpg", "photo2.jpghash", 1000)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if len(root.Children) != 3 || root.Children[2].Name != "photo1" || root.Children[2].Children[0].Hash != "photo2.jpghash" {
		t.Fatalf("Adding /photo1/photo2.jpg failed")
	}

	_, err = util.InsertFile(&root, "/photo1/subfolder/subfolder1/photo3.jpg", "photo3.jpghash", 999999999)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	if len(root.Children) != 3 || root.Children[2].Children[1].Name != "subfolder" || root.Children[2].Children[1].Children[0].Name != "subfolder1" || root.Children[2].Children[1].Children[0].Children[0].Name != "photo3.jpg" || root.Children[2].Children[1].Children[0].Children[0].Hash != "photo3.jpghash" {
		t.Fatalf("Adding /photo1/subfolder/subfolder1/photo3.jpg failed")
	}

	files := util.ListDir(&root, "/")
	if files == nil {
		t.Fatalf("List directory failed")
	}

	files = util.ListDir(&root, "/photo1")
	if files == nil {
		t.Fatalf("List directory failed")
	}
	if files[0].Name != "photo2.jpg" || files[1].Name != "subfolder" {
		t.Fatalf("Listed directory not correct")
	}

	fl := util.AddDir(&root, "/AddedFolder")
	if fl.Name != "AddedFolder" || fl.Type != "d" {
		t.Fatalf("Add dir failed")
	}

	fl = util.GetFileInfo(&root, "/photo1/subfolder/subfolder1/photo3.jpg")
	if fl == nil || fl.Name != "photo3.jpg" || fl.Type != "f" || fl.Size != 999999999 {
		t.Fatalf("Get File info failed")
	}

	fl = util.GetFileInfo(&root, "/photo2.jpg")
	if fl == nil || fl.Name != "photo2.jpg" || fl.Type != "f" || fl.Size != 100 {
		t.Fatalf("Get File info failed")
	}

	err = util.DeleteFile(&root, "/photo1/subfolder/subfolder1/photo3.jpg")
	if err != nil {
		t.Fatalf("File delete failed")
	}
	fl = util.GetFileInfo(&root, "/photo1/subfolder/subfolder1/photo3.jpg")
	if fl != nil {
		t.Fatalf("File exists after deleting")
	}
}

func TestFileMetaMap(t *testing.T) {
	root := util.NewDirTree()
	fl, err := util.InsertFile(&root, "/photo1.jpg", "photo1.jpghash", 10)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	fl.Meta["Key1"] = "Valuestring"
	fl.Meta["IntegerKey"] = 10
	if fl.Meta["Key1"].(string) != "Valuestring" || fl.Meta["IntegerKey"].(int) != 10 {
		t.Fatal("Map test failed")
	}
}
