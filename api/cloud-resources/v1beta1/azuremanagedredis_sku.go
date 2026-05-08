package v1beta1

// +kubebuilder:validation:Enum=Balanced_B1;Balanced_B3;Balanced_B5;Balanced_B10;Balanced_B20;Balanced_B50;Balanced_B100;Balanced_B150;Balanced_B250;Balanced_B350;Balanced_B500;Balanced_B700;Balanced_B1000;ComputeOptimized_X5;ComputeOptimized_X10;ComputeOptimized_X20;ComputeOptimized_X50;ComputeOptimized_X100;ComputeOptimized_X150;ComputeOptimized_X250;ComputeOptimized_X350;MemoryOptimized_E5;MemoryOptimized_E10;MemoryOptimized_E20;MemoryOptimized_E50;MemoryOptimized_E100;MemoryOptimized_E150;MemoryOptimized_E200;Flash_F300;Flash_F700;Flash_F1500
type AzureManagedRedisSKU string

const (
	AzureManagedRedisSKUBalancedB1    AzureManagedRedisSKU = "Balanced_B1"
	AzureManagedRedisSKUBalancedB3    AzureManagedRedisSKU = "Balanced_B3"
	AzureManagedRedisSKUBalancedB5    AzureManagedRedisSKU = "Balanced_B5"
	AzureManagedRedisSKUBalancedB10   AzureManagedRedisSKU = "Balanced_B10"
	AzureManagedRedisSKUBalancedB20   AzureManagedRedisSKU = "Balanced_B20"
	AzureManagedRedisSKUBalancedB50   AzureManagedRedisSKU = "Balanced_B50"
	AzureManagedRedisSKUBalancedB100  AzureManagedRedisSKU = "Balanced_B100"
	AzureManagedRedisSKUBalancedB150  AzureManagedRedisSKU = "Balanced_B150"
	AzureManagedRedisSKUBalancedB250  AzureManagedRedisSKU = "Balanced_B250"
	AzureManagedRedisSKUBalancedB350  AzureManagedRedisSKU = "Balanced_B350"
	AzureManagedRedisSKUBalancedB500  AzureManagedRedisSKU = "Balanced_B500"
	AzureManagedRedisSKUBalancedB700  AzureManagedRedisSKU = "Balanced_B700"
	AzureManagedRedisSKUBalancedB1000 AzureManagedRedisSKU = "Balanced_B1000"

	AzureManagedRedisSKUComputeOptimizedX5   AzureManagedRedisSKU = "ComputeOptimized_X5"
	AzureManagedRedisSKUComputeOptimizedX10  AzureManagedRedisSKU = "ComputeOptimized_X10"
	AzureManagedRedisSKUComputeOptimizedX20  AzureManagedRedisSKU = "ComputeOptimized_X20"
	AzureManagedRedisSKUComputeOptimizedX50  AzureManagedRedisSKU = "ComputeOptimized_X50"
	AzureManagedRedisSKUComputeOptimizedX100 AzureManagedRedisSKU = "ComputeOptimized_X100"
	AzureManagedRedisSKUComputeOptimizedX150 AzureManagedRedisSKU = "ComputeOptimized_X150"
	AzureManagedRedisSKUComputeOptimizedX250 AzureManagedRedisSKU = "ComputeOptimized_X250"
	AzureManagedRedisSKUComputeOptimizedX350 AzureManagedRedisSKU = "ComputeOptimized_X350"

	AzureManagedRedisSKUMemoryOptimizedE5   AzureManagedRedisSKU = "MemoryOptimized_E5"
	AzureManagedRedisSKUMemoryOptimizedE10  AzureManagedRedisSKU = "MemoryOptimized_E10"
	AzureManagedRedisSKUMemoryOptimizedE20  AzureManagedRedisSKU = "MemoryOptimized_E20"
	AzureManagedRedisSKUMemoryOptimizedE50  AzureManagedRedisSKU = "MemoryOptimized_E50"
	AzureManagedRedisSKUMemoryOptimizedE100 AzureManagedRedisSKU = "MemoryOptimized_E100"
	AzureManagedRedisSKUMemoryOptimizedE150 AzureManagedRedisSKU = "MemoryOptimized_E150"
	AzureManagedRedisSKUMemoryOptimizedE200 AzureManagedRedisSKU = "MemoryOptimized_E200"

	AzureManagedRedisSKUFlashF300  AzureManagedRedisSKU = "Flash_F300"
	AzureManagedRedisSKUFlashF700  AzureManagedRedisSKU = "Flash_F700"
	AzureManagedRedisSKUFlashF1500 AzureManagedRedisSKU = "Flash_F1500"
)
