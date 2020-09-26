package storage

import mgr "github.com/t-bfame/diago/internal/manager"

func AddTestInstance(test *mgr.TestInstance) error {
	return daoFactory.GetTestInstanceDao().AddTestInstance(test)
}

func GetTestInstanceByTestInstanceId(testInstanceId mgr.TestInstanceID) (*mgr.TestInstance, error) {
	return daoFactory.GetTestInstanceDao().GetTestInstanceByTestInstanceId(testInstanceId)
}
