// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package migrator

import (
	"log"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpAccActPathIn   string
	tpAccActPathOut  string
	tpAccActCfgIn    *config.CGRConfig
	tpAccActCfgOut   *config.CGRConfig
	tpAccActMigrator *Migrator
	tpAccountActions []*utils.TPAccountActions
)

var sTestsTpAccActIT = []func(t *testing.T){
	testTpAccActITConnect,
	testTpAccActITFlush,
	testTpAccActITPopulate,
	testTpAccActITMove,
	testTpAccActITCheckData,
}

func TestTpAccActMove(t *testing.T) {
	for _, stest := range sTestsTpAccActIT {
		t.Run("TestTpAccActMove", stest)
	}
}

func testTpAccActITConnect(t *testing.T) {
	var err error
	tpAccActPathIn = path.Join(*dataDir, "conf", "samples", "tutmongo")
	tpAccActCfgIn, err = config.NewCGRConfigFromFolder(tpAccActPathIn)
	if err != nil {
		t.Fatal(err)
	}
	tpAccActPathOut = path.Join(*dataDir, "conf", "samples", "tutmysql")
	tpAccActCfgOut, err = config.NewCGRConfigFromFolder(tpAccActPathOut)
	if err != nil {
		t.Fatal(err)
	}
	storDBIn, err := engine.ConfigureStorDB(tpAccActCfgIn.StorDBType, tpAccActCfgIn.StorDBHost,
		tpAccActCfgIn.StorDBPort, tpAccActCfgIn.StorDBName,
		tpAccActCfgIn.StorDBUser, tpAccActCfgIn.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	storDBOut, err := engine.ConfigureStorDB(tpAccActCfgOut.StorDBType,
		tpAccActCfgOut.StorDBHost, tpAccActCfgOut.StorDBPort, tpAccActCfgOut.StorDBName,
		tpAccActCfgOut.StorDBUser, tpAccActCfgOut.StorDBPass,
		config.CgrConfig().StorDBMaxOpenConns,
		config.CgrConfig().StorDBMaxIdleConns,
		config.CgrConfig().StorDBConnMaxLifetime,
		config.CgrConfig().StorDBCDRSIndexes)
	if err != nil {
		log.Fatal(err)
	}
	tpAccActMigrator, err = NewMigrator(nil, nil, tpAccActCfgIn.DataDbType,
		tpAccActCfgIn.DBDataEncoding, storDBIn, storDBOut, tpAccActCfgIn.StorDBType, nil,
		tpAccActCfgIn.DataDbType, tpAccActCfgIn.DBDataEncoding, nil,
		tpAccActCfgIn.StorDBType, false, false, false, false, false)
	if err != nil {
		log.Fatal(err)
	}
}

func testTpAccActITFlush(t *testing.T) {
	if err := tpAccActMigrator.storDBIn.Flush(
		path.Join(tpAccActCfgIn.DataFolderPath, "storage", tpAccActCfgIn.StorDBType)); err != nil {
		t.Error(err)
	}

	if err := tpAccActMigrator.storDBOut.Flush(
		path.Join(tpAccActCfgOut.DataFolderPath, "storage", tpAccActCfgOut.StorDBType)); err != nil {
		t.Error(err)
	}
}

func testTpAccActITPopulate(t *testing.T) {
	tpAccountActions = []*utils.TPAccountActions{
		&utils.TPAccountActions{
			TPid:          "TPAcc",
			LoadId:        "ID",
			Tenant:        "cgrates.org",
			Account:       "1001",
			ActionPlanId:  "PREPAID_10",
			AllowNegative: true,
			Disabled:      false,
		},
	}
	if err := tpAccActMigrator.storDBIn.SetTPAccountActions(tpAccountActions); err != nil {
		t.Error("Error when setting TpAccountActions ", err.Error())
	}
	currentVersion := engine.CurrentStorDBVersions()
	err := tpAccActMigrator.storDBOut.SetVersions(currentVersion, false)
	if err != nil {
		t.Error("Error when setting version for TpAccountActions ", err.Error())
	}
}

func testTpAccActITMove(t *testing.T) {
	err, _ := tpAccActMigrator.Migrate([]string{utils.MetaTpAccountActions})
	if err != nil {
		t.Error("Error when migrating TpAccountActions ", err.Error())
	}
}

func testTpAccActITCheckData(t *testing.T) {
	filter := &utils.TPAccountActions{TPid: tpAccountActions[0].TPid}
	result, err := tpAccActMigrator.storDBOut.GetTPAccountActions(filter)
	if err != nil {
		t.Error("Error when getting TpAccountActions ", err.Error())
	}
	if !reflect.DeepEqual(tpAccountActions[0], result[0]) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(tpAccountActions[0]), utils.ToJSON(result[0]))
	}
	result, err = tpAccActMigrator.storDBIn.GetTPAccountActions(filter)
	if err != utils.ErrNotFound {
		t.Error(err)
	}
}