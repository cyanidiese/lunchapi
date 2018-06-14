package controllers

import (
	"lunchapi/app/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/revel/revel"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open("mysql", "root:@/lunch?parseTime=true&charset=utf8")
	if err != nil {
		panic(err)
	}
	DB.LogMode(true)

	if !DB.HasTable(&models.Translation{}){ DB.CreateTable(&models.Translation{}) }
	if !DB.HasTable(&models.Office{}){ DB.CreateTable(&models.Office{}) }
	if !DB.HasTable(&models.Image{}){ DB.CreateTable(&models.Image{}) }
	if !DB.HasTable(&models.Role{}){ DB.CreateTable(&models.Role{}) }
	if !DB.HasTable(&models.User{}){ DB.CreateTable(&models.User{}) }
	if !DB.HasTable(&models.Category{}){ DB.CreateTable(&models.Category{}) }
	if !DB.HasTable(&models.Dish{}){ DB.CreateTable(&models.Dish{}) }
	if !DB.HasTable(&models.Menu{}){ DB.CreateTable(&models.Menu{}) }
	if !DB.HasTable(&models.MenuItem{}){ DB.CreateTable(&models.MenuItem{})



		DbSeedInitialData()
	}

	DB.AutoMigrate(
		&models.Translation{},
		&models.Office{},
		&models.Image{},
		&models.Role{},
		&models.User{},
		&models.Category{},
		&models.Dish{},
		&models.Menu{},
		&models.MenuItem{})
}

func DbSeedInitialData(){


	//======================  ROLES  ===============================

	providerRole := models.Role{Name: "provider", Title: "Provider"}
	DB.Create(&providerRole)
	DB.Save(&providerRole)

	//=====================  OFFICES  ================================

	office := models.Office{
		Title: models.Translation{En: "Paris Office", Ua: "Офіс на Парижці", Ru: "Парижский офис"},
		Phone: "5345235435",
		Lat: 49.432193,
		Lng: 32.083675,
		Address: "Some long and interesting address",
	}
	DB.Create(&office)
	DB.Save(&office)

	providerOffice := models.Office{
		Title: models.Translation{En: "Prov Office", Ua: "Офіс Провайдера", Ru: "Провайдерский офис"},
		Phone: "5345235435",
		Lat: 49.432193,
		Lng: 32.083675,
		Address: "Some long and interesting address",
		IsProvider: true,
	}
	DB.Create(&providerOffice)
	DB.Save(&providerOffice)

	//=====================  USERS  ================================

	admin := models.User{
		FirstName: "Test",
		LastName: "Admin",
		Email: "admin@test.lunch",
		Token: AuthRandToken(),
		Office: office,
		Role: models.Role{Name: "admin", Title: "Admin"},
		Image: models.Image{Guid: "6c57aa9f-1725-4417-b115-b1c98f6d4ed7"},
		Timezone: "Europe/Kiev",
	}

	DB.Create(&admin)
	DB.Save(&admin)

	provider := models.User{
		FirstName: "Test",
		LastName: "Provider",
		Email: "provider@test.lunch",
		Token: AuthRandToken(),
		IsProvider: true,
		Office: providerOffice,
		Role: providerRole,
		Image: models.Image{Guid: "6c57aa9f-1725-4417-b115-b1c98f6d4ed7"},
		Timezone: "Europe/Kiev",
	}

	DB.Create(&provider)
	DB.Save(&provider)

	provider2 := models.User{
		FirstName: "Second",
		LastName: "Provider",
		Email: "provider2@test.lunch",
		Token: AuthRandToken(),
		IsProvider: true,
		Office: providerOffice,
		Role: providerRole,
		Image: models.Image{Guid: "6c57aa9f-1725-4417-b115-b1c98f6d4ed7"},
		Timezone: "Europe/Kiev",
	}

	DB.Create(&provider2)
	DB.Save(&provider2)

	master := models.User{
		FirstName: "Test",
		LastName: "Master",
		Email: "master@test.lunch",
		Token: AuthRandToken(),
		Office: office,
		Role: models.Role{Name: "master", Title: "Master"},
		Image: models.Image{Guid: "6c57aa9f-1725-4417-b115-b1c98f6d4ed7"},
		Timezone: "Europe/Kiev",
	}

	DB.Create(&master)
	DB.Save(&master)

	//======================  DISHES  ===============================

	firstDish := models.Dish{
		Name: models.Translation{ En: "Tasty First", Ua: "Перша страва", Ru: "Первое блюдо" },
		Description: models.Translation{ En: "Veeeery Tasty First", Ua: "Перше завжди найкраще", Ru: "Первое всегда вкусное" },
		Weight: 453,
		Calories: 2432,
		Price: 3.14,
		ProviderId: provider.Id,
		Provider: provider,
		Category: models.Category{ Title: models.Translation{ En: "First", Ua: "Перше", Ru: "Первое" } },
		Images: []models.Image{{Guid: "97696ba1-f65f-4b6e-a3b1-d622cfacc3e2"}},
	}
	DB.Create(&firstDish)
	DB.Save(&firstDish)

	secondDish := models.Dish{
		Name: models.Translation{ En: "Some Second", Ua: "Друга страва", Ru: "Второе блюдо" },
		Description: models.Translation{ En: "Smelly Second Dish", Ua: "Друге завжди пахне", Ru: "Второе всегда пахнет" },
		Weight: 1436,
		Calories: 656,
		Price: 121.42,
		ProviderId: provider.Id,
		Provider: provider,
		Category: models.Category{ Title: models.Translation{ En: "Second", Ua: "Друге", Ru: "Второе" } },
		Images: []models.Image{{Guid: "61099d67-5a6f-4959-9db0-5698884e12f5"},{Guid: "d1ec0ae1-164a-405f-b312-70f4227299ea"}},
	}
	DB.Create(&secondDish)
	DB.Save(&secondDish)

	dessertCategory := models.Category{ Title: models.Translation{ En: "Dessert", Ua: "Десерт", Ru: "Дессерт" } }
	DB.Create(&dessertCategory)
	DB.Save(&dessertCategory)

	dessertDish := models.Dish{
		Name: models.Translation{ En: "Ice Cream", Ua: "Морозивко", Ru: "Мороженка" },
		Description: models.Translation{ En: "So tasty and so sweet!", Ua: "Таке смачне що нівмагату", Ru: "Ощень вкусненько" },
		Weight: 236,
		Calories: 623656,
		Price: 11.87,
		Provider: provider,
		ProviderId: provider.Id,
		Category: dessertCategory,
		CategoryId: dessertCategory.Id,
		Images: []models.Image{{Guid: "842d2066-a380-41d9-a6ef-02380a942155"}},
	}
	DB.Create(&dessertDish)
	DB.Save(&dessertDish)

	secondProviderDish := models.Dish{
		Name: models.Translation{ En: "Second Dessert", Ua: "Другий десерт", Ru: "Второй десерт" },
		Description: models.Translation{ En: "Lorem ipsum", Ua: "Лорем іпсум", Ru: "Лорем ипсум" },
		Weight: 8658,
		Calories: 764,
		Price: 34.45,
		Provider: provider2,
		ProviderId: provider2.Id,
		CategoryId: dessertCategory.Id,
		Category: dessertCategory,
		Images: []models.Image{{Guid: "d1ec0ae1-164a-405f-b312-70f4227299ea"}},
	}
	DB.Create(&secondProviderDish)
	DB.Save(&secondProviderDish)

	//=====================================================
}

type GormController struct {
	*revel.Controller
	DB *gorm.DB
}