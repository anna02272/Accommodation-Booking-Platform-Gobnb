package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type RateHost struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Host        *User              `bson:"host" json:"host"`
	Guest       *User              `bson:"guest" json:"guest"`
	DateAndTime primitive.DateTime `bson:"date-and-time" json:"date-and-time"`
	Rating      Rating             `bson:"rating" json:"rating"`
}
type RateAccommodation struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Accommodation string             `bson:"accommodationID" json:"accommodationID"`
	Guest         *User              `bson:"guest" json:"guest"`
	DateAndTime   primitive.DateTime `bson:"date-and-time" json:"date-and-time"`
	Rating        Rating             `bson:"rating" json:"rating"`
}
type RateAccommodationRec struct {
	ID            string ` json:"id"`
	Accommodation string ` json:"accommodationID"`
	Guest         string `json:"guest"`
	Rating        int    `json:"rating"`
}
type Rating int

const (
	one   = 1
	two   = 2
	three = 3
	four  = 4
	five  = 5
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username string             `bson:"username" json:"username"`
	Email    string             `bson:"email" json:"email" validate:"required,email"`
}
type NeoUser struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
}
type NeoUsers []*NeoUser

type Accommodation struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Location string             `bson:"location" json:"location"`
}

type ReservationByGuest struct {
	ReservationIdTimeCreated string    `json:"reservation_id_time_created"`
	GuestId                  string    `json:"guest_id"`
	AccommodationId          string    `json:"accommodation_id"`
	AccommodationName        string    `json:"accommodation_name"`
	AccommodationLocation    string    `json:"accommodation_location"`
	AccommodationHostId      string    `json:"accommodation_host_id"`
	CheckInDate              time.Time `json:"check_in_date"`
	CheckOutDate             time.Time `json:"check_out_date"`
	NumberOfGuests           int       `json:"number_of_guests"`
}

type UserResponse struct {
	Message string `json:"message"`
	User    struct {
		ID       primitive.ObjectID `json:"id"`
		Username string             `json:"username"`
		Email    string             `json:"email"`
		Name     string             `json:"name"`
		Lastname string             `json:"lastname"`
		Address  Address            `json:"address"`
		Age      int                `json:"age"`
		Gender   string             `json:"gender"`
		UserRole UserRole           `json:"userRole"`
	} `json:"user"`
}
type AccommodationRec struct {
	ID       string `json:"_id"`
	HostId   string `json:"host_id"`
	Name     string `json:"accommodation_name"`
	Location string `json:"accommodation_location"`
	//Amenities map[string]bool    `json:"accommodation_amenities"`
	MinGuests int  `json:"accommodation_min_guests"`
	MaxGuests int  `json:"accommodation_max_guests"`
	Active    bool `json:"accommodation_active"`
}

type AccommodationResponse struct {
	Message       string `json:"message"`
	Accommodation struct {
		ID        primitive.ObjectID `json:"_id"`
		HostId    string             `json:"host_id"`
		Name      string             ` json:"accommodation_name"`
		Location  string             `json:"accommodation_location"`
		Amenities map[string]bool    `json:"accommodation_amenities"`
		MinGuests int                `json:"accommodation_min_guests"`
		MaxGuests int                `json:"accommodation_max_guests"`
		Active    bool               `json:"accommodation_active"`
	} `json:"accommodation"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
}
type UserRole string

const (
	Guest = "Guest"
	Host  = "Host"
)

func ConvertToDomainUser(userResponse UserResponse) User {
	return User{
		ID:       userResponse.User.ID,
		Username: userResponse.User.Username,
		Email:    userResponse.User.Email,
	}
}
func ConvertToDomainAccommodation(accommodationResponse AccommodationResponse) Accommodation {
	return Accommodation{
		ID:       accommodationResponse.Accommodation.ID,
		Name:     accommodationResponse.Accommodation.Name,
		Location: accommodationResponse.Accommodation.Location,
	}
}
