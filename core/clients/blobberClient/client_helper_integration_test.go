package blobberClient

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/0chain/blobber/code/go/0chain.net/core/common"

	"github.com/0chain/gosdk/core/zcncrypto"

	"gorm.io/gorm"
)

func randString(n int) string {

	const hexLetters = "abcdef0123456789"

	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteByte(hexLetters[rand.Intn(len(hexLetters))])
	}
	return sb.String()
}

type TestDataController struct {
	db *gorm.DB
}

func NewTestDataController(db *gorm.DB) *TestDataController {
	return &TestDataController{db: db}
}

// ClearDatabase deletes all data from all tables
func (c *TestDataController) ClearDatabase() error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	_, err = tx.Exec("truncate allocations cascade")
	if err != nil {
		return err
	}

	_, err = tx.Exec("truncate reference_objects cascade")
	if err != nil {
		return err
	}

	_, err = tx.Exec("truncate commit_meta_txns cascade")
	if err != nil {
		return err
	}

	_, err = tx.Exec("truncate collaborators cascade")
	if err != nil {
		return err
	}

	_, err = tx.Exec("truncate allocation_changes cascade")
	if err != nil {
		return err
	}

	_, err = tx.Exec("truncate allocation_connections cascade")
	if err != nil {
		return err
	}

	_, err = tx.Exec("truncate write_markers cascade")
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddGetAllocationTestData() error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, repairer_id, is_immutable)
VALUES ('exampleId' ,'exampleTransaction','exampleOwnerId','exampleOwnerPublicKey',` + fmt.Sprint(expTime) + `,'examplePayerId', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddGetFileMetaDataTestData() error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, repairer_id, is_immutable)
VALUES ('exampleId' ,'exampleTransaction','exampleOwnerId','exampleOwnerPublicKey',` + fmt.Sprint(expTime) + `,'examplePayerId', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO reference_objects (id, allocation_id, path_hash,lookup_hash,type,name,path,hash,custom_meta,content_hash,merkle_root,actual_file_hash,mimetype,write_marker,thumbnail_hash, actual_thumbnail_hash)
VALUES (1234,'exampleId','exampleId:examplePath','exampleId:examplePath','f','filename','examplePath','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash');
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO commit_meta_txns (ref_id,txn_id)
VALUES (1234,'someTxn');
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO collaborators (ref_id, client_id)
VALUES (1234, 'someClient');
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddGetFileStatsTestData(allocationTx, pubKey string) error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, repairer_id, is_immutable)
VALUES ('exampleId' ,'` + allocationTx + `','exampleOwnerId','` + pubKey + `',` + fmt.Sprint(expTime) + `,'examplePayerId', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO reference_objects (id, allocation_id, path_hash,lookup_hash,type,name,path,hash,custom_meta,content_hash,merkle_root,actual_file_hash,mimetype,write_marker,thumbnail_hash, actual_thumbnail_hash)
VALUES (1234,'exampleId','exampleId:examplePath','exampleId:examplePath','f','filename','examplePath','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash');
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddListEntitiesTestData(allocationTx, pubkey string) error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, repairer_id, is_immutable)
VALUES ('exampleId' ,'` + allocationTx + `','exampleOwnerId','` + pubkey + `',` + fmt.Sprint(expTime) + `,'examplePayerId', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO reference_objects (id, allocation_id, path_hash,lookup_hash,type,name,path,hash,custom_meta,content_hash,merkle_root,actual_file_hash,mimetype,write_marker,thumbnail_hash, actual_thumbnail_hash)
VALUES (1234,'exampleId','exampleId:examplePath','exampleId:examplePath','f','filename','examplePath','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash');
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddGetObjectPathTestData(allocationTx, pubKey string) error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, repairer_id, is_immutable)
VALUES ('exampleId' ,'` + allocationTx + `','exampleOwnerId','` + pubKey + `',` + fmt.Sprint(expTime) + `,'examplePayerId', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddGetReferencePathTestData(allocationTx, pubkey string) error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, repairer_id, is_immutable)
VALUES ('exampleId' ,'` + allocationTx + `','exampleOwnerId','` + pubkey + `',` + fmt.Sprint(expTime) + `,'examplePayerId', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddGetObjectTreeTestData(allocationTx, pubkey string) error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, repairer_id, is_immutable)
VALUES ('exampleId' ,'` + allocationTx + `','exampleOwnerId','` + pubkey + `',` + fmt.Sprint(expTime) + `,'examplePayerId', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO reference_objects (id, allocation_id, path_hash,lookup_hash,type,name,path,hash,custom_meta,content_hash,merkle_root,actual_file_hash,mimetype,write_marker,thumbnail_hash, actual_thumbnail_hash)
VALUES (1234,'exampleId','exampleId:examplePath','exampleId:examplePath','d','root','/','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash');
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func GeneratePubPrivateKey(t *testing.T) (string, string, zcncrypto.SignatureScheme) {

	signScheme := zcncrypto.NewSignatureScheme("bls0chain")
	wallet, err := signScheme.GenerateKeys()
	if err != nil {
		t.Fatal(err)
	}
	keyPair := wallet.Keys[0]

	_ = signScheme.SetPrivateKey(keyPair.PrivateKey)
	return keyPair.PublicKey, keyPair.PrivateKey, signScheme
}

func (c *TestDataController) AddCommitTestData(allocationTx, pubkey, clientId, wmSig string, now common.Timestamp) error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, blobber_size, allocation_root, repairer_id, is_immutable)
VALUES ('exampleId' ,'` + allocationTx + `','` + clientId + `','` + pubkey + `',` + fmt.Sprint(expTime) + `,'examplePayerId', 99999999, '/', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO allocation_connections (connection_id, allocation_id, client_id, size, status)
VALUES ('connection_id' ,'exampleId','` + clientId + `', 1337, 1);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO allocation_changes (id, connection_id, operation, size, input)
VALUES (1 ,'connection_id','rename', 1200, '{"allocation_id":"exampleId","path":"/some_file","new_name":"new_name"}');
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO write_markers(prev_allocation_root, allocation_root, status, allocation_id, size, client_id, signature, blobber_id, timestamp, connection_id, client_key)
VALUES ('/', '/', 2,'exampleId', 1337, '` + clientId + `','` + wmSig + `','blobber_id', ` + fmt.Sprint(now) + `, 'connection_id', '` + pubkey + `');
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO reference_objects (id, allocation_id, path_hash,lookup_hash,type,name,path,hash,custom_meta,content_hash,merkle_root,actual_file_hash,mimetype,write_marker,thumbnail_hash, actual_thumbnail_hash, parent_path)
VALUES 
(1234,'exampleId','exampleId:examplePath','exampleId:examplePath','d','root','/','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash','/'),
(123,'exampleId','exampleId:examplePath','exampleId:examplePath','f','some_file','/some_file','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash','/');
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddAttributesTestData(allocationTx, pubkey, clientId string) error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, blobber_size, allocation_root, repairer_id, is_immutable)
VALUES ('exampleId' ,'` + allocationTx + `','` + clientId + `','` + pubkey + `',` + fmt.Sprint(expTime) + `,'examplePayerId', 99999999, '/', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO allocation_connections (connection_id, allocation_id, client_id, size, status)
VALUES ('connection_id' ,'exampleId','` + clientId + `', 1337, 1);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO reference_objects (id, allocation_id, path_hash,lookup_hash,type,name,path,hash,custom_meta,content_hash,merkle_root,actual_file_hash,mimetype,write_marker,thumbnail_hash, actual_thumbnail_hash, parent_path)
VALUES 
(1234,'exampleId','exampleId:examplePath','exampleId:examplePath','d','root','/','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash','/'),
(123,'exampleId','exampleId:examplePath','exampleId:examplePath','f','some_file','/some_file','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash','/');
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddCopyObjectData(allocationTx, pubkey, clientId string) error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, blobber_size, allocation_root, repairer_id, is_immutable)
VALUES ('exampleId' ,'` + allocationTx + `','` + clientId + `','` + pubkey + `',` + fmt.Sprint(expTime) + `,'examplePayerId', 99999999, '/', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO allocation_connections (connection_id, allocation_id, client_id, size, status)
VALUES ('connection_id' ,'exampleId','` + clientId + `', 1337, 1);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO reference_objects (id, allocation_id, path_hash,lookup_hash,type,name,path,hash,custom_meta,content_hash,merkle_root,actual_file_hash,mimetype,write_marker,thumbnail_hash, actual_thumbnail_hash, parent_path)
VALUES 
(1234,'exampleId','exampleId:examplePath','exampleId:examplePath','d','root','/copy','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash','/'),
(123,'exampleId','exampleId:examplePath','exampleId:examplePath','f','some_file','/some_file','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash','/');
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddRenameTestData(allocationTx, pubkey, clientId string) error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, blobber_size, allocation_root, repairer_id, is_immutable)
VALUES ('exampleId' ,'` + allocationTx + `','` + clientId + `','` + pubkey + `',` + fmt.Sprint(expTime) + `,'examplePayerId', 99999999, '/', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO allocation_connections (connection_id, allocation_id, client_id, size, status)
VALUES ('connection_id' ,'exampleId','` + clientId + `', 1337, 1);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO reference_objects (id, allocation_id, path_hash,lookup_hash,type,name,path,hash,custom_meta,content_hash,merkle_root,actual_file_hash,mimetype,write_marker,thumbnail_hash, actual_thumbnail_hash, parent_path)
VALUES 
(1234,'exampleId','exampleId:examplePath','exampleId:examplePath','d','root','/','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash','/'),
(123,'exampleId','exampleId:examplePath','exampleId:examplePath','f','some_file','/some_file','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash','/');
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddDownloadTestData(allocationTx, pubkey, clientId, wmSig string, now common.Timestamp) error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, blobber_size, allocation_root, repairer_id, is_immutable)
VALUES ('exampleId' ,'` + allocationTx + `','` + clientId + `','` + pubkey + `',` + fmt.Sprint(expTime) + `,'examplePayerId', 99999999, '/', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO allocation_connections (connection_id, allocation_id, client_id, size, status)
VALUES ('connection_id' ,'exampleId','` + clientId + `', 1337, 1);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO allocation_changes (id, connection_id, operation, size, input)
VALUES (1 ,'connection_id','rename', 1200, '{"allocation_id":"exampleId","path":"/some_file","new_name":"new_name"}');
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO reference_objects (id, allocation_id, path_hash,lookup_hash,type,name,path,hash,custom_meta,content_hash,merkle_root,actual_file_hash,mimetype,write_marker,thumbnail_hash, actual_thumbnail_hash, parent_path)
VALUES 
(1234,'exampleId','exampleId:examplePath','exampleId:examplePath','d','root','/','someHash','customMeta','contentHash','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash','/'),
(123,'exampleId','exampleId:examplePath','exampleId:examplePath','f','some_file','/some_file','someHash','customMeta','tmpMonWenMyFile','merkleRoot','actualFileHash','mimetype','writeMarker','thumbnailHash','actualThumbnailHash','/');
`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *TestDataController) AddUploadTestData(allocationTx, pubkey, clientId string, wmSig string, now common.Timestamp) error {
	var err error
	var tx *sql.Tx
	defer func() {
		if err != nil {
			if tx != nil {
				errRollback := tx.Rollback()
				if errRollback != nil {
					log.Println(errRollback)
				}
			}
		}
	}()

	db, err := c.db.DB()
	if err != nil {
		return err
	}

	tx, err = db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	expTime := time.Now().Add(time.Hour * 100000).UnixNano()

	_, err = tx.Exec(`
INSERT INTO allocations (id, tx, owner_id, owner_public_key, expiration_date, payer_id, blobber_size, allocation_root, repairer_id, is_immutable)
VALUES ('exampleId' ,'` + allocationTx + `','` + clientId + `','` + pubkey + `',` + fmt.Sprint(expTime) + `,'examplePayerId', 99999999, '/', 'repairer_id', false);
`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	INSERT INTO allocation_connections (connection_id, allocation_id, client_id, size, status)
	VALUES ('connection_id' ,'exampleId','` + clientId + `', 1337, 1);
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	INSERT INTO allocation_changes (id, connection_id, operation, size, input)
	VALUES (1 ,'connection_id','rename', 1200, '{"allocation_id":"exampleId","path":"/some_file","new_name":"new_name"}');
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	INSERT INTO write_markers(prev_allocation_root, allocation_root, status, allocation_id, size, client_id, signature, blobber_id, timestamp, connection_id, client_key)
	VALUES ('/', '/', 2,'exampleId', 1337, '` + clientId + `','` + wmSig + `','blobber_id', ` + fmt.Sprint(now) + `, 'connection_id', '` + pubkey + `');
	`)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
