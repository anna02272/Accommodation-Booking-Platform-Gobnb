package utils

//func CreateToken(ttl time.Duration, payload interface{}, privateKey string) (string, error) {
//	decodedPrivateKey, err := base64.StdEncoding.DecodeString(privateKey)
//	if err != nil {
//		return "", fmt.Errorf("could not decode key: %w", err)
//	}
//	key, err := jwt.ParseRSAPrivateKeyFromPEM(decodedPrivateKey)
//
//	if err != nil {
//		return "", fmt.Errorf("create: parse key: %w", err)
//	}
//
//	now := time.Now().UTC()
//
//	claims := make(jwt.MapClaims)
//	claims["sub"] = payload
//	claims["exp"] = now.Add(ttl).Unix()
//	claims["iat"] = now.Unix()
//	claims["nbf"] = now.Unix()
//
//	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
//
//	if err != nil {
//		return "", fmt.Errorf("create: sign token: %w", err)
//	}
//
//	return token, nil
//}
//
//func ValidateToken(token string, publicKey string) (interface{}, error) {
//	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
//	if err != nil {
//		return nil, fmt.Errorf("could not decode: %w", err)
//	}
//
//	key, err := jwt.ParseRSAPublicKeyFromPEM(decodedPublicKey)
//
//	if err != nil {
//		return "", fmt.Errorf("validate: parse key: %w", err)
//	}
//
//	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
//		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
//			return nil, fmt.Errorf("unexpected method: %s", t.Header["alg"])
//		}
//		return key, nil
//	})
//
//	if err != nil {
//		return nil, fmt.Errorf("validate: %w", err)
//	}
//
//	claims, ok := parsedToken.Claims.(jwt.MapClaims)
//	if !ok || !parsedToken.Valid {
//		return nil, fmt.Errorf("validate: invalid token")
//	}
//
//	return claims["sub"], nil
//}
//func DeserializeUser(userService services.UserService) gin.HandlerFunc {
//	return func(ctx *gin.Context) {
//		var access_token string
//		cookie, err := ctx.Cookie("access_token")
//
//		authorizationHeader := ctx.Request.Header.Get("Authorization")
//		fields := strings.Fields(authorizationHeader)
//
//		if len(fields) != 0 && fields[0] == "Bearer" {
//			access_token = fields[1]
//		} else if err == nil {
//			access_token = cookie
//		}
//
//		if access_token == "" {
//			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "You are not logged in"})
//			return
//		}
//
//		config, _ := config.LoadConfig(".")
//		sub, err := ValidateToken(access_token, config.AccessTokenPublicKey)
//		if err != nil {
//			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": err.Error()})
//			return
//		}
//
//		user, err := userService.FindUserById(fmt.Sprint(sub))
//		if err != nil {
//			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "The user belonging to this token no logger exists"})
//			return
//		}
//
//		ctx.Set("currentUser", user)
//		ctx.Next()
//	}
//}
