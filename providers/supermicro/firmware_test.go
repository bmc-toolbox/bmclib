package supermicro

//func TestDeviceModel(t *testing.T) {
//
//	l := logrus.New()
//	l.Level = logrus.DebugLevel
//	logger := logrusr.New(l)
//	c := NewClient("10.251.153.157", "ADMIN", "XWMCYBJEPL", logger)
//
//	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
//	defer cancel()
//
//	if err := c.Open(ctx); err != nil {
//		log.Fatal("login error" + err.Error())
//	}
//
//	defer c.Close(ctx)
//
//	ok, err := c.x12().redfish.BmcReset(ctx, "GracefulRestart")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Println(ok)
//	//	running, err := c.x12().bmcFirmwareInstallRunning(ctx)
//	//	if err != nil {
//	//		log.Fatal(err)
//	//	}
//
//	fmt.Println("hello")
//	// fmt.Println(running)
//}
