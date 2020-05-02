#!/bin/bash

EXEC_EXPERIMENT_PATH='/root/execExperiment.sh'

OVERPROV_PERCENTAGES='0 50 100'

MIN=30
MAX=60
for OVERPROV_PERC in ${OVERPROV_PERCENTAGES}
do
	echo 'MIN = '${MIN}' /MAX = '${MAX}' /OVERPROVISIONING PERCENTAGE = '${OVERPROV_PERC}
	${EXEC_EXPERIMENT_PATH} ${MIN} ${MAX} ${OVERPROV_PERC}
done
