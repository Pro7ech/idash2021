import struct
import numpy as np
import tensorflow as tf
from random import shuffle
from PIL import Image
import matplotlib
import matplotlib.cm as cm
from scipy.special import softmax
import matplotlib.pyplot as plt
from tensorflow.keras import layers, initializers, regularizers, optimizers, callbacks


class Regression():
    def __init__(self, X: np.ndarray, Y: np.ndarray):

        self.model = tf.keras.Sequential()

        self.X = X
        self.Y = Y

        self.kernel_initializer = 'he_normal'
 
        self.callbacks = callbacks.EarlyStopping(monitor='loss', patience=50)
        
        self.history = None
        self.layer_added = False
        self.compiled = False
        self.trained = False

    def add(self,
            units,
            activation=None,
            use_bias=True,
            kernel_initializer='glorot_uniform',
            bias_initializer='zeros',
            kernel_regularizer=None,
            bias_regularizer=None,
            activity_regularizer=None,
        ):

        if self.layer_added:
            raise Exception("layer already added")
        
        self.model.add(tf.keras.layers.Dense(
                units=units,
                activation=activation,
                use_bias=use_bias,
                kernel_initializer=kernel_initializer,
                bias_initializer=bias_initializer,
                kernel_regularizer=kernel_regularizer,
                bias_regularizer=bias_regularizer,
                activity_regularizer=activity_regularizer,
                input_shape=(len(self.X[0]), )
                )
            )


        self.layer_added = True

    def compile(self, optimizer="rmsprop", loss=None, metrics=None):

        if not self.layer_added:
            raise Exception("add a layer before compiling")

        if self.compiled:
            raise Exception("model is already compiled")
        
        self.model.compile(
            optimizer=optimizer,
            loss=loss,
            metrics=metrics,
            )
        self.compiled = True 

    def set_callback(self, callbacks:callbacks):
        self.callbacks = callbacks
    
    def fit(self, epochs:int, batch_size:int=32, verbose:int=2, validation_split:float=0.5):

        if not self.compiled:
            raise Exception("compile model before calling fit")
        
        self.history = self.model.fit(
            self.X,
            self.Y,
            epochs=epochs,
            batch_size=batch_size,
            verbose=verbose,
            validation_split=validation_split,
            callbacks = [self.callbacks]
        )

        self.trained = True

    def test_for_encrypted_categorical_evaluation(self, log_scale_samples:int, log_scale_model: int):

        if not self.trained:
            raise Exception("fit model before testing accuracy with encryption")

        X = self.X[:]
        Y = self.Y[:]
 
        # Quantizes samples
        X *= (1<<log_scale_samples)
        X = np.round(X)
        X /= (1<<log_scale_samples)

        # Loads model
        layer = self.model.layers[0]
        weights = layer.get_weights()
        W = weights[0][:]
        b = weights[1][:]

        # Quantizes model
        W *= (1<<log_scale_model)
        W = np.round(W)
        W /= (1<<log_scale_model)

        b *= (1<<log_scale_model)
        b = np.round(b)
        b /= (1<<log_scale_model)

        err = 0
        minDiff = 10
        minL = 0
        maxL = 0

        for i in range(len(X)):

            # Linear prediction
            L = X[i] @ W + b

            L = L.tolist()

            #Get the index of the predicted classe
            idx = L.index(np.max(L))

            # Sort
            L.sort()
            L = L[::-1] #revert order

            # Minimum value
            minL = min(minL, L[-1])

            #Maximum value
            maxL = max(maxL, L[0])

            #Difference between the to most likely classes
            minDiff = min(minDiff, L[0]-L[1])

            #Increments by 1 if the prediction is wrong
            if idx != Y[i].tolist().index(1):
                err += 1

        print("classification errors :", err)
        print("minimum absolute value before activation", minL)
        print("maximum absolute value before activation", maxL)
        print("minimum distance between two highest categories", minDiff)

    def plot_history(self):

        if self.history is None:
            raise Exception("fit model before plotting history")

        history = self.history
        
        plt.rc('font', size=32)    
        plt.rc('axes', titlesize=32) 
        plt.rc('axes', labelsize=24)
        plt.rc('xtick', labelsize=16)
        plt.rc('ytick', labelsize=16)
        plt.rc('legend', fontsize=24) 
        plt.rc('figure', titlesize=16)
        plt.rcParams.update()

        # summarize history for accuracy
        plt.plot(history.history['categorical_accuracy'])
        plt.plot(history.history['val_categorical_accuracy'])
        plt.title('Model Accuracy')
        plt.ylabel('Accuracy')
        plt.xlabel('Epoch')
        plt.legend(['train', 'valid'], loc='lower right')
        plt.savefig('accuracy.png')
        plt.clf()
        
        # summarize history for loss
        plt.plot(history.history['loss'])
        plt.plot(history.history['val_loss'])
        plt.title('Model Loss')
        plt.ylabel('Loss')
        plt.xlabel('Epoch')
        plt.legend(['train', 'valid'], loc='upper right')
        plt.savefig('loss.png')
        plt.clf()
    
    def save(self):

        if not self.trained:
            raise Exception("fit model before saving it")
    
        #Save both bias and weights in bianry and npy format
        layer = self.model.layers[0]
        weights = layer.get_weights()
        data = []
        flattened_weights = weights[0].flatten()
        for j in range(len(flattened_weights)):
            data += struct.pack('d', flattened_weights[j])

        with open("weights_layer_{}".format(0), "wb") as f:
            f.write(bytearray(data))
            f.close()
        np.save('weights_layer_{}.npy'.format(0), weights[0])
        
        data = []
        flattened_bias = weights[1].flatten()
        for j in range(len(flattened_bias)):
            data += struct.pack('d', flattened_bias[j])

        with open("bias_layer_{}".format(0), "wb") as f:
            f.write(bytearray(data))
            f.close()
        np.save('bias_layer_{}.npy'.format(0), weights[1])

def load_samples(shuffle_samples):

    hash_size = 0
    with open('params.binary', "rb") as f:
        data = f.read()
        hash_size = int(data[0])**2
    
    X = []
    Y = []
    nb_samples = 0
    with open('X.binary', "rb") as f:
        data = f.read()
        nb_samples = len(data)//(hash_size*8)
        for i in range(nb_samples):
            tmp = [0.0 for i in range(hash_size)]
            buff = data[(i*hash_size)*8:(i+1)*hash_size*8]
            for j in range(hash_size):
                    tmp[j] = struct.unpack('d', buff[j*8:(j+1)*8])[0]
            X += [tmp]

    with open('Y.binary', "rb") as f:
        data = f.read()
        for i in data:
            tmp = [0, 0, 0, 0]
            tmp[int(i)] = 1
            Y += [tmp]

    if shuffle_samples:
        available = [i for i in range(nb_samples)]
        shuffle(available)
        X_shuffled = []
        Y_shuffled = []    
        for i in available:
            X_shuffled += [X[i]]
            Y_shuffled += [Y[i]]
        return np.array(X_shuffled), np.array(Y_shuffled)
    else:
        return np.array(X), np.array(Y)

if __name__ == "__main__":
    X, Y = load_samples(True)

    regression = Regression(X, Y)
    regression.add(units=4, activation="softmax", kernel_regularizer=regularizers.l1_l2(l1=1e-6, l2=1e-6))
    regression.compile(optimizer="adam", loss='mean_squared_error', metrics='categorical_accuracy')
    regression.fit(epochs=256, batch_size=32, verbose=2, validation_split=0.5)
    regression.test_for_encrypted_categorical_evaluation(log_scale_samples=16, log_scale_model=12)
    regression.plot_history()
    regression.save()
    
