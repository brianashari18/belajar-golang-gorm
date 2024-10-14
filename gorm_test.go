package golang_gorm

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"strconv"
	"testing"
)

func OpenConnection() *gorm.DB {
	dialect := mysql.Open("root:Pokemon18*@tcp(localhost:3306)/golang_gorm?charset=utf8mb4&parseTime=True&loc=Local")
	db, err := gorm.Open(dialect, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	return db

}

var db = OpenConnection()

func TestOpenConnection(t *testing.T) {
	assert.NotNil(t, db)
}

func TestExecuteSQL(t *testing.T) {
	err := db.Exec("insert into sample(id, name) values(?, ?)", "1", "Brian").Error
	assert.Nil(t, err)

	err = db.Exec("insert into sample(id, name) values(?, ?)", "2", "Anashari").Error
	assert.Nil(t, err)

	err = db.Exec("insert into sample(id, name) values(?, ?)", "3", "Sari").Error
	assert.Nil(t, err)

	err = db.Exec("insert into sample(id, name) values(?, ?)", "4", "Puyol").Error
	assert.Nil(t, err)

	err = db.Exec("insert into sample(id, name) values(?, ?)", "5", "Celox").Error
	assert.Nil(t, err)
}

type Sample struct {
	Id   string
	Name string
}

func TestQuerySQL(t *testing.T) {
	var sample Sample
	err := db.Raw("select id, name from sample where id = ?", "1").Scan(&sample).Error
	assert.Nil(t, err)

	var samples []Sample
	err = db.Raw("select id, name from sample").Scan(&samples).Error
	assert.Nil(t, err)
	assert.Equal(t, 5, len(samples))
}

func TestSqlRow(t *testing.T) {
	rows, err := db.Raw("select id, name from sample where id = ?", "1").Rows()
	assert.Nil(t, err)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		assert.Nil(t, err)
	}(rows)

	var samples []Sample
	for rows.Next() {
		var id string
		var name string
		err = rows.Scan(&id, &name)
		assert.Nil(t, err)

		samples = append(samples, Sample{Id: id, Name: name})
	}
	assert.Equal(t, 1, len(samples))
}

func TestScanRow(t *testing.T) {
	rows, err := db.Raw("select id, name from sample").Rows()
	assert.Nil(t, err)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		assert.Nil(t, err)
	}(rows)

	var samples []Sample
	for rows.Next() {
		err = db.ScanRows(rows, &samples)
		assert.Nil(t, err)
	}
	assert.Equal(t, 5, len(samples))
}

func TestCreateUsers(t *testing.T) {
	user := User{
		ID:       "1",
		Password: "rahasia",
		Name: Name{
			FirstName: "Brian",
			LastName:  "Anashari",
		},
		Information: "This information will be ignored",
	}

	tx := db.Create(&user)
	assert.Nil(t, tx.Error)
	assert.Equal(t, int64(1), tx.RowsAffected)
}

func TestBatchInsert(t *testing.T) {
	var users []User
	for i := 2; i < 10; i++ {
		user := User{
			ID:       strconv.Itoa(i),
			Password: "rahasia",
			Name: Name{
				FirstName: "User",
				LastName:  strconv.Itoa(i),
			},
			Information: "This information will be ignored",
		}
		users = append(users, user)
	}
	tx := db.Create(&users)
	assert.Nil(t, tx.Error)
	assert.Equal(t, int64(8), tx.RowsAffected)
}

func TestTransactionSuccess(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := db.Create(&User{ID: "11", Password: "rahasia", Name: Name{FirstName: "User 11"}}).Error
		if err != nil {
			return err
		}

		err = db.Create(&User{ID: "12", Password: "rahasia", Name: Name{FirstName: "User 12"}}).Error
		if err != nil {
			return err
		}

		err = db.Create(&User{ID: "13", Password: "rahasia", Name: Name{FirstName: "User 13"}}).Error
		if err != nil {
			return err
		}

		return nil
	})

	assert.Nil(t, err)
}

func TestTransactionRollback(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&User{ID: "16", Password: "rahasia", Name: Name{FirstName: "User 16"}}).Error
		if err != nil {
			return err
		}

		err = tx.Create(&User{ID: "12", Password: "rahasia", Name: Name{FirstName: "User 12"}}).Error
		if err != nil {
			return err
		}

		return nil
	})

	assert.NotNil(t, err)
}

func TestManualTransactionSuccess(t *testing.T) {
	tx := db.Begin()
	defer tx.Rollback()

	err := tx.Create(&User{ID: "17", Password: "rahasia", Name: Name{FirstName: "User 17"}}).Error
	assert.Nil(t, err)

	err = tx.Create(&User{ID: "18", Password: "rahasia", Name: Name{FirstName: "User 18"}}).Error
	assert.Nil(t, err)

	if err == nil {
		tx.Commit()
	}
}

func TestManualTransactionRollback(t *testing.T) {
	tx := db.Begin()
	defer tx.Rollback()

	err := tx.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&User{ID: "17", Password: "rahasia", Name: Name{FirstName: "User 17"}}).Error
		if err != nil {
			return err
		}

		err = tx.Create(&User{ID: "20", Password: "rahasia", Name: Name{FirstName: "User 20"}}).Error
		if err != nil {
			return err
		}

		return nil
	})

	assert.NotNil(t, err)

	if err == nil {
		tx.Commit()
	}
}

func TestQuerySingleObject(t *testing.T) {
	user := User{}
	err := db.First(&user).Error
	assert.Nil(t, err)
	assert.Equal(t, "1", user.ID)

	user = User{}
	err = db.Last(&user).Error
	assert.Nil(t, err)
	assert.Equal(t, "9", user.ID)
}

func TestQuerySingleObjectInlineCondition(t *testing.T) {
	user := User{}
	err := db.First(&user, "id = ?", "5").Error
	assert.Nil(t, err)
	assert.Equal(t, "5", user.ID)

	user = User{}
	err = db.Take(&user, "id = ?", "4").Error
	assert.Nil(t, err)
	assert.Equal(t, "4", user.ID)
}

func TestQueryAllObject(t *testing.T) {
	var users []User
	err := db.Find(&users, "id in ?", []string{"1", "2"}).Error
	assert.Nil(t, err)
	assert.Equal(t, 2, len(users))
}

func TestQueryWhere(t *testing.T) {
	var user User
	err := db.Where("first_name like ?", "%User%").Where("id = ?", "5").Find(&user).Error
	assert.Nil(t, err)
	assert.Equal(t, "5", user.ID)
}

func TestQueryOr(t *testing.T) {
	var users []User
	err := db.Where("first_name like ?", "%User%").Or("password = ?", "rahasia").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 18, len(users))
}

func TestQueryNot(t *testing.T) {
	var users []User
	err := db.Not("first_name like ?", "%User%").Where("password = ?", "rahasia").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
}

func TestSelectFields(t *testing.T) {
	var users []User
	err := db.Select("id, first_name").Find(&users).Error
	assert.Nil(t, err)

	for _, user := range users {
		assert.NotNil(t, user.ID)
		assert.NotEqual(t, "", user.Name.FirstName)
	}

	assert.Equal(t, 18, len(users))
}

func TestStructCondition(t *testing.T) {
	userCondition := User{
		Name: Name{
			FirstName: "User 10",
		},
	}

	var users []User
	err := db.Where(userCondition).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
}

func TestMapCondition(t *testing.T) {
	userCondition := map[string]interface{}{
		"middle_name": "",
		"first_name":  []string{"User 10", "User 11"},
	}

	var users []User
	err := db.Where(userCondition).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 2, len(users))
}

func TestOrderLimitOffset(t *testing.T) {
	var users []User
	err := db.Order("id asc, first_name desc").Limit(5).Offset(5).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 5, len(users))
}

type UserResponse struct {
	ID        string
	FirstName string
	LastName  string
}

func TestQueryNonModel(t *testing.T) {
	var users []UserResponse
	err := db.Model(&User{}).Select("id, first_name, last_name").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 18, len(users))
}

func TestUpdate(t *testing.T) {
	var user User
	err := db.Where("id = ?", 10).Find(&user).Error
	assert.Nil(t, err)

	user.Name.FirstName = "Sari"
	user.Name.LastName = "Puyol"
	user.Password = "newpassword"
	err = db.Save(&user).Error
	assert.Nil(t, err)

	db.Where("id =?", 10).Where("password = ?", "newpassword").Find(&user)
	assert.Equal(t, "10", user.ID)
	assert.Equal(t, "newpassword", user.Password)
	assert.Equal(t, "Sari", user.Name.FirstName)
}

func TestUpdateSelectionColumns(t *testing.T) {
	err := db.Model(&User{}).Where("id = ?", 11).Updates(map[string]interface{}{
		"first_name": "Celox",
		"last_name":  "Dusk",
	}).Error
	assert.Nil(t, err)

	err = db.Model(&User{}).Where("id = ?", 12).Update("password", "hidden").Error
	assert.Nil(t, err)

	db.Where("id = ?", 13).Updates(User{
		Name: Name{
			FirstName: "Anas",
			LastName:  "Hari",
		},
	})
}

func TestAutoIncrement(t *testing.T) {
	for i := 0; i < 10; i++ {
		userLog := UserLog{
			UserId: "1",
			Action: "Test Action",
		}

		err := db.Create(&userLog).Error
		assert.Nil(t, err)
		assert.NotEqual(t, 0, userLog.ID)

		fmt.Println(userLog.ID)
	}
}

func TestCreateOrUpdate(t *testing.T) {
	userLog := UserLog{
		UserId: "1",
		Action: "Test Action",
	}

	err := db.Save(&userLog).Error
	assert.Nil(t, err)

	userLog.UserId = "2"
	err = db.Save(&userLog).Error
	assert.Nil(t, err)
}

func TestCreateOrUpdateNonAutoIncrement(t *testing.T) {
	user := User{
		ID:   "99",
		Name: Name{FirstName: "User 99"},
	}

	err := db.Save(&user).Error
	assert.Nil(t, err)

	user.Name = Name{FirstName: "User 99 updated"}
	err = db.Save(&user).Error
	assert.Nil(t, err)
}

func TestConflict(t *testing.T) {
	user := User{
		ID:   "88",
		Name: Name{FirstName: "User 88"},
	}

	err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&user).Error
	assert.Nil(t, err)
}

func TestDelete(t *testing.T) {
	var user User
	err := db.Take(&user, "id = ?", "99").Error
	assert.Nil(t, err)

	err = db.Delete(&user).Error
	assert.Nil(t, err)

	err = db.Delete(&User{}, "id = ?", "88").Error
	assert.Nil(t, err)

	err = db.Where("id = ?", "19").Delete(&User{}).Error
	assert.Nil(t, err)
}

func TestSoftDelete(t *testing.T) {
	todo := Todo{
		UserId:      "1",
		Title:       "1",
		Description: "Todo 1",
	}
	err := db.Create(&todo).Error
	assert.Nil(t, err)

	err = db.Delete(&todo).Error
	assert.Nil(t, err)
	assert.NotNil(t, todo.DeletedAt)

	var todos []Todo
	err = db.Find(&todo).Error
	assert.Nil(t, err)
	assert.Equal(t, 0, len(todos))
}

func TestUnscoped(t *testing.T) {
	var todo Todo
	err := db.Unscoped().First(&todo, "id = ?", "3").Error
	assert.Nil(t, err)

	err = db.Unscoped().Delete(&todo).Error
	assert.Nil(t, err)

	var todos []Todo
	err = db.Unscoped().Find(&todos).Error
	assert.Nil(t, err)
}

func TestLock(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		var user User
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&user, "id = ?", "1").Error
		if err != nil {
			return err
		}

		user.Name.FirstName = "User 1 updated"
		user.Password = "newupdate"
		err = tx.Save(&user).Error
		return err
	})
	assert.Nil(t, err)
}

func TestCreateWallet(t *testing.T) {
	wallet := Wallet{
		ID:      "1",
		UserId:  "1",
		Balance: 1000000000,
	}

	err := db.Create(&wallet).Error
	assert.Nil(t, err)
}

func TestRetrieveRelation(t *testing.T) {
	var user User
	err := db.Model(User{}).Preload("Wallet").Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)
	assert.Equal(t, "1", user.ID)
	assert.Equal(t, "1", user.Wallet.ID)
}

func TestRetrieveRelationJoin(t *testing.T) {
	var user User
	err := db.Model(User{}).Joins("Wallet").Take(&user, "users.id = ?", "1").Error
	assert.Nil(t, err)
	assert.Equal(t, "1", user.ID)
	assert.Equal(t, "1", user.Wallet.ID)
}

func TestAutoCreateUpdate(t *testing.T) {
	user := User{
		ID:       "20",
		Name:     Name{FirstName: "User 20"},
		Password: "rahasia",
		Wallet: Wallet{
			ID:      "20",
			UserId:  "20",
			Balance: 1000000000,
		},
	}

	err := db.Create(&user).Error
	assert.Nil(t, err)
}

func TestSkipAutoCreateUpdate(t *testing.T) {
	user := User{
		ID:       "21",
		Name:     Name{FirstName: "User 21"},
		Password: "rahasia",
		Wallet: Wallet{
			ID:      "21",
			UserId:  "21",
			Balance: 1000000000,
		},
	}

	err := db.Omit(clause.Associations).Create(&user).Error
	assert.Nil(t, err)
}

func TestUserAndAddresses(t *testing.T) {
	user := User{
		ID:       "23",
		Name:     Name{FirstName: "User 23"},
		Password: "rahasia",
		Wallet: Wallet{
			ID:      "23",
			UserId:  "23",
			Balance: 1000000000,
		},
		Addresses: []Address{
			{
				UserId:  "23",
				Address: "Jalan A",
			},
			{
				UserId:  "23",
				Address: "Jalan B",
			},
		},
	}

	err := db.Create(&user).Error
	assert.Nil(t, err)
}

func TestPreloadJoinOneToMany(t *testing.T) {
	var users []User
	err := db.Model(&User{}).Preload("Addresses").
		Joins("Wallet").Find(&users).Error
	assert.Nil(t, err)
}

func TestTakePreloadJoinOneToMany(t *testing.T) {
	var user User
	err := db.Model(&User{}).Preload("Addresses").
		Joins("Wallet").Take(&user, "users.id = ?", "20").Error
	assert.Nil(t, err)
}

func TestBelongsTo(t *testing.T) {
	fmt.Println("Preload")
	var addresses []Address
	err := db.Model(&Address{}).Preload("User").Find(&addresses).Error
	assert.Nil(t, err)
	assert.Equal(t, 4, len(addresses))

	fmt.Println("Join")
	err = db.Model(&Address{}).Joins("User").Find(&addresses).Error
	assert.Nil(t, err)
	assert.Equal(t, 4, len(addresses))
}

func TestBelongsToWallet(t *testing.T) {
	fmt.Println("Preload")
	var wallets []Wallet
	err := db.Model(&Wallet{}).Preload("User").Find(&wallets).Error
	assert.Nil(t, err)

	fmt.Println("Join")
	err = db.Model(&Wallet{}).Joins("User").Find(&wallets).Error
	assert.Nil(t, err)
}

func TestCreateManyToMany(t *testing.T) {
	product := Product{
		ID:    "P001",
		Name:  "Contoh Produk",
		Price: 1000000,
	}

	err := db.Create(&product).Error
	assert.Nil(t, err)

	err = db.Table("user_like_product").Create(map[string]interface{}{
		"user_id":    "1",
		"product_id": "P001",
	}).Error
	assert.Nil(t, err)

	err = db.Table("user_like_product").Create(map[string]interface{}{
		"user_id":    "2",
		"product_id": "P001",
	}).Error
	assert.Nil(t, err)
}

func TestPreloadManyToMany(t *testing.T) {
	var product Product
	err := db.Preload("LikedByUsers").Take(&product, "id = ?", "P001").Error
	assert.Nil(t, err)
	assert.Equal(t, 3, len(product.LikedByUsers))
}

func TestPreloadManyToManyUser(t *testing.T) {
	var user User
	err := db.Preload("LikeProducts").Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)
	assert.Equal(t, 1, len(user.LikeProducts))
}

func TestAssociationFind(t *testing.T) {
	var product Product
	err := db.Take(&product, "id = ?", "P001").Error
	assert.Nil(t, err)

	var users []User
	err = db.Model(&product).Where("first_name like ?", "User%").Association("LikedByUsers").Find(&users)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(users))
}

func TestAssociationAdd(t *testing.T) {
	var product Product
	err := db.Take(&product, "id = ?", "P001").Error
	assert.Nil(t, err)

	var user User
	err = db.Take(&user, "id = ?", "3").Error
	assert.Nil(t, err)

	err = db.Model(&product).Association("LikedByUsers").Append(&user)
	assert.Nil(t, err)
}

func TestAssociationReplace(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		var user User
		err := db.Take(&user, "id = ?", "1").Error
		if err != nil {
			return err
		}

		wallet := Wallet{
			ID:      "01",
			UserId:  user.ID,
			Balance: 1000000,
		}

		err = db.Model(&user).Association("Wallet").Replace(&wallet)
		return err
	})
	assert.Nil(t, err)
}

func TestAssociationDelete(t *testing.T) {
	var product Product
	err := db.Take(&product, "id = ?", "P001").Error
	assert.Nil(t, err)

	var user User
	err = db.Take(&user, "id = ?", "3").Error
	assert.Nil(t, err)

	err = db.Model(&product).Association("LikedByUsers").Delete(&user)
	assert.Nil(t, err)
}

func TestAssociationClear(t *testing.T) {
	var product Product
	err := db.Take(&product, "id = ?", "P001").Error
	assert.Nil(t, err)

	var user User
	err = db.Take(&user, "id = ?", "3").Error
	assert.Nil(t, err)

	err = db.Model(&product).Association("LikedByUsers").Clear()
	assert.Nil(t, err)
}

func TestPreloadingWithCondition(t *testing.T) {
	var user User
	err := db.Preload("Wallet", "balance > ?", 100).Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)

	fmt.Println(user)
}

func TestPreloadingNested(t *testing.T) {
	var wallet Wallet
	err := db.Preload("User.Addresses").Take(&wallet, "id = ?", "22").Error
	assert.Nil(t, err)

	fmt.Println(wallet)
	fmt.Println(wallet.User)
	fmt.Println(wallet.User.Addresses)
}

func TestPreloadingAll(t *testing.T) {
	var user User
	err := db.Preload(clause.Associations).Take(&user, "id = ?", "1").Error
	assert.Nil(t, err)
}
